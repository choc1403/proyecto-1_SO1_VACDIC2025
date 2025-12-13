package functions

import (
	"encoding/json"
	"fmt"
	"log"
	"so1-daemon/database"
	"so1-daemon/utils"
	"so1-daemon/var_const"
	"strings"
	"time"
)

func DecideAndAct(containers []var_const.ProcProcess) {
	// Build docker pid map
	dmap, err := GetDockerPidMap()
	if err != nil {
		log.Printf("Warning: cannot get docker map: %v", err)
	}

	// classify containers discovered (prefer mapping via docker inspect)
	type CInfo struct {
		Proc   var_const.ProcProcess
		Docker var_const.DockerInfo
	}
	var detected []CInfo
	for _, p := range containers {

		// Dentro de DecideAndAct, dentro del bucle for _, p := range containers
		if d, ok := dmap[p.Pid]; ok {
			// Caso 1: Proceso principal, usa info de Docker
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			// Caso 2: Proceso que no es el principal (Podría ser un SHIM)

			// Intenta encontrar el ID del contenedor en la línea de comandos del shim
			if p.Name == "containerd-shim" {

				// **Lógica para extraer el Container ID del p.Cmdline**
				// Podrías usar expresiones regulares o manipulación de strings
				// buscando la cadena "-id " y deteniéndote en el siguiente espacio.

				containerID := ExtractContainerID(p.Cmdline)

				if containerID != "" {
					// **Buscar la imagen real usando el Container ID**
					dockerInfo, err := GetDockerInfoByID(containerID) // Nueva función auxiliar

					if err == nil {
						detected = append(detected, CInfo{Proc: p, Docker: dockerInfo})
						// ¡Ahora c.Docker.Image tendrá el nombre real (ej: 'low_img')!
						continue
					}
				}
			}

			// Caso 3: Proceso genérico (o shim fallido)
			detected = append(detected, CInfo{
				Proc: p,
				Docker: var_const.DockerInfo{
					ContainerID: "",
					Image:       p.Cmdline, // Usar Cmdline por defecto
					Pid:         p.Pid,
					Name:        p.Name,
				},
			})
		}

	}

	// count low/high based on image naming heuristic: image contains "low" or "high"
	lowCount := 0
	highCount := 0
	for _, c := range detected {

		img := strings.ToLower(c.Docker.Image)

		isLow := strings.Contains(img, "low_img")
		isHighCPU := strings.Contains(img, "high_cpu_img")
		isHighRAM := strings.Contains(img, "high_mem_img")

		if isHighCPU || isHighRAM {
			highCount++
		} else if isLow {
			lowCount++
		}

	}

	log.Printf("Informacion: low=%d high=%d", lowCount, highCount)

	totalJiffies, _ := ReadTotalJiffies()
	now := time.Now()
	type decisionCandidate struct {
		C   CInfo
		Mem float64
		Cpu float64
	}
	var candidates []decisionCandidate

	for _, c := range detected {
		memf, _ := utils.ParseMemPct(c.Proc.MemPct)
		// compute cpu via /proc/<pid>/stat and /proc/stat
		procTime, err := ReadProcPidTime(c.Proc.Pid)
		if err != nil {
			// couldn't read stat; set cpu 0
			procTime = 0
		}
		cpuPct := CalcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)
		candidates = append(candidates, decisionCandidate{C: c, Mem: memf, Cpu: cpuPct})

		// Save record in DB
		log.Println("Guardando en la base de datos...")
		database.InsertContainerRecord(c.Docker.ContainerID, c.Proc.Pid, c.Docker.Image, cpuPct, memf)
	}
	for _, cand := range candidates {
		img := strings.ToLower(cand.C.Docker.Image)

		isLow := strings.Contains(img, "low_img")
		isHighCPU := strings.Contains(img, "high_cpu_img")
		isHighRAM := strings.Contains(img, "high_mem_img")

		shouldKill := false
		reason := ""

		log.Println("CPU: ", cand.Cpu, " RAM: ", cand.Mem)

		if isHighCPU && cand.Cpu > var_const.CPU_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("cpu %.2f > %.2f", cand.Cpu, var_const.CPU_THRESHOLD)
		}
		if isHighRAM && cand.Mem > var_const.MEM_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("mem %.2f > %.2f", cand.Mem, var_const.MEM_THRESHOLD)
		}
		if isLow && (cand.Cpu > var_const.CPU_THRESHOLD || cand.Mem > var_const.MEM_THRESHOLD) {
			shouldKill = true
			reason = "low container exceeded threshold"
		}
		if shouldKill {
			if cand.C.Docker.ContainerID == "" {
				// if not a docker container, skip deletion (can't)
				log.Printf("Candidate pid %d not a docker container or no id available, skip deletion", cand.C.Proc.Pid)
				continue
			}
			// don't delete grafana
			if strings.Contains(strings.ToLower(cand.C.Docker.Image), "grafana") || strings.Contains(strings.ToLower(cand.C.Docker.Name), "grafana") {
				log.Printf("Skipping deletion of grafana container %s", cand.C.Docker.ContainerID)
				continue
			}
			if isHighCPU || isHighRAM {
				if highCount <= var_const.MIN_HIGH_CONTAINERS {
					log.Printf("Would delete %s but would violate MIN_HIGH_CONTAINERS (%d)", cand.C.Docker.ContainerID, var_const.MIN_HIGH_CONTAINERS)
					continue
				}
			} else if isLow {
				if lowCount <= var_const.MIN_LOW_CONTAINERS {
					log.Printf(
						"Would delete %s but would violate MIN_LOW_CONTAINERS (%d)",
						cand.C.Docker.ContainerID,
						var_const.MIN_LOW_CONTAINERS,
					)
					continue
				}

			} else {
				if lowCount <= var_const.MIN_LOW_CONTAINERS {
					log.Printf(
						"Would delete %s (unclassified image) but would violate MIN_LOW_CONTAINERS (%d)",
						cand.C.Docker.ContainerID,
						var_const.MIN_LOW_CONTAINERS,
					)
					continue
				}
			}

			log.Printf("Deleting container %s due to %s (cpu=%.2f mem=%.2f)", cand.C.Docker.ContainerID, reason, cand.Cpu, cand.Mem)
			out, err := utils.RunCommand("docker", "rm", "-f", cand.C.Docker.ContainerID)
			if err != nil {
				log.Printf("Failed to remove container %s: %v | out: %s", cand.C.Docker.ContainerID, err, out)
			} else {
				database.InsertDeletion(cand.C.Docker.ContainerID, reason)
				// adjust counts
				if isHighCPU || isHighRAM {
					highCount--
				} else {
					lowCount--
				}
			}
		}
	}

}

func ProcessOnce() error {
	// read sys metrics
	sysB, err := utils.ReadProcFile(var_const.PROC_SYS)
	if err != nil {
		return fmt.Errorf("read sys proc: %v", err)
	}
	sysB = utils.SanitizeJSON(sysB)
	var sys var_const.ProcSys
	if err := json.Unmarshal(sysB, &sys); err != nil {
		return fmt.Errorf("parse sys json: %v", err)
	}
	database.InsertSysMetrics(sys.MemTotalKb, sys.MemFreeKb, sys.MemUsedKb)
	database.InsertProcessCount(len(sys.Processes))

	// read cont info
	contB, err := utils.ReadProcFile(var_const.PROC_CONT)
	if err != nil {
		return fmt.Errorf("read cont proc: %v", err)
	}
	contB = utils.SanitizeJSON(contB)
	var cont var_const.ProcCont
	if err := json.Unmarshal(contB, &cont); err != nil {
		return fmt.Errorf("parse cont json: %v", err)
	}

	// analyze and act
	DecideAndAct(cont.Containers)
	return nil
}
