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

// DecideAndAct analiza el consumo de recursos de los contenedores detectados
// y toma decisiones automáticas (por ejemplo, eliminar contenedores)
// según políticas de CPU, memoria y reglas de balance mínimo.
//
// Flujo general:
// 1) Obtiene el mapeo PID ↔ Contenedor Docker
// 2) Clasifica procesos como contenedores reales, shims o genéricos
// 3) Cuenta contenedores low / high según la imagen
// 4) Calcula uso de CPU y memoria
// 5) Aplica reglas de decisión y ejecuta acciones (docker rm)
func DecideAndAct(containers []var_const.ProcProcess) {

	// 1. Construcción del mapa PID → Información Docker
	// Obtiene los contenedores activos usando docker inspect
	// El mapa permite relacionar un PID con su contenedor real
	dmap, err := GetDockerPidMap()
	if err != nil {
		log.Printf("Advertencia: no se puede obtener el mapa de Docker: %v", err)
	}

	// 2. Clasificación de procesos detectados
	// Une información del kernel (/proc) con Docker
	type CInfo struct {
		Proc   var_const.ProcProcess
		Docker var_const.DockerInfo
	}
	var detected []CInfo
	for _, p := range containers {

		// Caso 1: PID corresponde directamente a un contenedor Docker
		if d, ok := dmap[p.Pid]; ok {
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			// Caso 2: Proceso intermedio (containerd-shim)
			// Se intenta extraer el Container ID desde la línea de comandos
			if p.Name == "containerd-shim" {

				containerID := ExtractContainerID(p.Cmdline)

				if containerID != "" {
					// **Buscar la imagen real usando el Container ID**
					dockerInfo, err := GetDockerInfoByID(containerID) // Función auxiliar

					if err == nil {
						detected = append(detected, CInfo{Proc: p, Docker: dockerInfo})
						continue
					}
				}
			}

			// Caso 3: Proceso genérico (no identificado como contenedor real)
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

	// 3. Conteo de contenedores LOW / HIGH
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

	log.Printf("Información: Bajo consumo=%d Alto Consumo=%d", lowCount, highCount)

	// 4. Preparación para cálculo de CPU y memoria

	totalJiffies, _ := ReadTotalJiffies()
	now := time.Now()
	type decisionCandidate struct {
		C   CInfo
		Mem float64
		Cpu float64
	}
	var candidates []decisionCandidate

	for _, c := range detected {

		// Parseo del porcentaje de memoria

		memf, _ := utils.ParseMemPct(c.Proc.MemPct)

		// Caso: proceso no asociado a un contenedor Docker
		if c.Docker.ContainerID == "" {
			log.Printf("Skipping cgroup read for non-docker PID %d", c.Proc.Pid)

			procTime, err := ReadProcPidTime(c.Proc.Pid)

			if err != nil {
				log.Printf("Warning: failed to read proc time for PID %d: %v", c.Proc.Pid, err)
				procTime = 0
			}
			cpuPct := CalcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)

			candidates = append(candidates, decisionCandidate{C: c, Mem: memf, Cpu: cpuPct})
			database.InsertContainerRecord(c.Docker.ContainerID, c.Proc.Pid, c.Docker.Image, cpuPct, memf)
			continue
		}

		// --- NUEVA LECTURA DEL CGROUP ---
		// Esto lee el tiempo total de CPU en nanosegundos (la fuente de datos de Docker).
		procTime, err := ReadCgroupCpuTime(c.Docker.ContainerID)

		if err != nil {
			log.Printf("Advertencia: no se ha podido leer el tiempo de CPU del grupo de control para %s: %v. Se establece la CPU en 0.", c.Docker.ContainerID, err)
			procTime = 0
		}

		// 2. Usar este valor para el cálculo.
		cpuPct := CalcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)
		candidates = append(candidates, decisionCandidate{C: c, Mem: memf, Cpu: cpuPct})

		// Registrar en base de datos
		database.InsertContainerRecord(c.Docker.ContainerID, c.Proc.Pid, c.Docker.Image, cpuPct, memf)
	}

	// 5. Evaluación de reglas y acciones

	for _, cand := range candidates {
		img := strings.ToLower(cand.C.Docker.Image)

		isLow := strings.Contains(img, "low_img")
		isHighCPU := strings.Contains(img, "high_cpu_img")
		isHighRAM := strings.Contains(img, "high_mem_img")

		shouldKill := false
		reason := ""
		if isHighCPU {
			log.Println("CPU: ", cand.Cpu, " RAM: ", cand.Mem, " RASONAMIENTO CPU: ", isHighCPU && cand.Cpu > var_const.CPU_THRESHOLD,
				" RASONAMIENTO RAM: ", isHighRAM && cand.Mem > var_const.MEM_THRESHOLD)

		}
		// Reglas de eliminación
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
			reason = "El contenedor bajo ha superado el umbral."
		}
		if shouldKill {
			if cand.C.Docker.ContainerID == "" {
				// if not a docker container, skip deletion (can't)
				log.Printf("El candidato pid %d no es un contenedor Docker o no hay ningún ID disponible, omitir la eliminación.", cand.C.Proc.Pid)
				continue
			}
			// don't delete grafana
			if strings.Contains(strings.ToLower(cand.C.Docker.Image), "grafana") || strings.Contains(strings.ToLower(cand.C.Docker.Name), "grafana") {
				log.Printf("Omitiendo la eliminación del contenedor grafana %s", cand.C.Docker.ContainerID)
				continue
			}
			if isHighCPU || isHighRAM {
				if highCount <= var_const.MIN_HIGH_CONTAINERS {
					log.Printf("Se eliminaría %s, pero se infringiría MIN_HIGH_CONTAINERS (%d)", cand.C.Docker.ContainerID, var_const.MIN_HIGH_CONTAINERS)
					continue
				}
			} else if isLow {
				if lowCount <= var_const.MIN_LOW_CONTAINERS {
					log.Printf(
						"Se eliminaría %s, pero se infringiría MIN_LOW_CONTAINERS (%d)",
						cand.C.Docker.ContainerID,
						var_const.MIN_LOW_CONTAINERS,
					)
					continue
				}

			} else {
				if lowCount <= var_const.MIN_LOW_CONTAINERS {
					log.Printf(
						"Se eliminaría %s (imagen sin clasificar), pero se infringiría MIN_LOW_CONTAINERS (%d).",
						cand.C.Docker.ContainerID,
						var_const.MIN_LOW_CONTAINERS,
					)
					continue
				}
			}

			log.Printf("Eliminación del contenedor %s debido a %s (cpu=%.2f mem=%.2f)", cand.C.Docker.ContainerID, reason, cand.Cpu, cand.Mem)
			out, err := utils.RunCommand("docker", "rm", "-f", cand.C.Docker.ContainerID)
			if err != nil {
				log.Printf("No se pudo eliminar el contenedor %s: %v | salida: %s", cand.C.Docker.ContainerID, err, out)
			} else {
				database.InsertDeletion(cand.C.Docker.ContainerID, reason)

				if isHighCPU || isHighRAM {
					highCount--
				} else {
					lowCount--
				}
			}
		}
	}

}

// ProcessOnce ejecuta un ciclo completo de monitoreo del sistema.
//
// La función realiza las siguientes tareas:
// 1) Lee métricas generales del sistema desde /proc/sysinfo
// 2) Registra métricas de memoria y cantidad de procesos en la base de datos
// 3) Lee información de contenedores desde /proc/continfo
// 4) Analiza el estado de los contenedores y ejecuta acciones correctivas
//
// Esta función es invocada periódicamente por el daemon principal
// mediante un ticker (por ejemplo, cada 20 segundos).
func ProcessOnce() error {

	// 1. Lectura de métricas generales del sistema

	// Lee el archivo /proc/sysinfo generado por el módulo del kernel
	sysB, err := utils.ReadProcFile(var_const.PROC_SYS)
	if err != nil {
		return fmt.Errorf("leer sys proc: %v", err)
	}

	// Limpia el JSON generado por el kernel (comas finales, formatos no estándar)
	sysB = utils.SanitizeJSON(sysB)

	// Estructura destino para deserializar la información del sistema
	var sys var_const.ProcSys

	// Convierte el JSON en una estructura Go
	if err := json.Unmarshal(sysB, &sys); err != nil {
		return fmt.Errorf("analizar sys json: %v", err)
	}

	// Inserta métricas de memoria del sistema en la base de datos
	database.InsertSysMetrics(
		sys.MemTotalKb,
		sys.MemFreeKb,
		sys.MemUsedKb,
	)

	// Registra la cantidad total de procesos activos
	database.InsertProcessCount(len(sys.Processes))

	// 2. Lectura de información de contenedores

	// Lee el archivo /proc/continfo generado por el módulo del kernel
	contB, err := utils.ReadProcFile(var_const.PROC_CONT)
	if err != nil {
		return fmt.Errorf("leer cont proc: %v", err)
	}

	// Limpia el JSON de salida del módulo del kernel
	contB = utils.SanitizeJSON(contB)

	// Estructura destino para deserializar la información de contenedores
	var cont var_const.ProcCont

	// Convierte el JSON en la estructura Go correspondiente
	if err := json.Unmarshal(contB, &cont); err != nil {
		return fmt.Errorf("analizar cont json: %v", err)
	}

	// 3. Análisis y toma de decisiones

	// Analiza el consumo de recursos de los contenedores
	DecideAndAct(cont.Containers)

	return nil
}
