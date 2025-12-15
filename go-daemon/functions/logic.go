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
	/*
		type decisionCandidate struct {
			C   CInfo
			Mem float64
			Cpu float64
		}*/
	//var candidates []decisionCandidate

	for _, c := range detected {

		// 1. Obtener la hora del proceso directamente desde /proc/<pid>/stat.
		// Esto garantiza que obtendrás la medición de alta precisión del kernel.
		procTime, err := ReadProcPidTime(c.Proc.Pid)

		if err != nil {
			log.Printf("Warning: failed to read proc time for PID %d: %v", c.Proc.Pid, err)
			procTime = 0
		}

		// 2. Usar este valor para el cálculo.
		cpuPct := CalcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)

		// ... (El resto del log)
		// Puedes loggear el valor de procTime real para verificar que se esté actualizando
		log.Println("DEL PROC (JSON): ", c.Proc.ProcJiffies, " PROC TIME (kernel): ", procTime, " Obtenido: ", totalJiffies, "cpuPct: ", cpuPct)
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
