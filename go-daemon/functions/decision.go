package functions

import (
	"fmt"
	"log"
	"strings"
	"time"
	"strconv"

	"proyecto1/go-daemon/database"
)

/*
   Este archivo contiene la lógica de decisiones:
   - evaluar CPU y MEM
   - respetar límites mínimos de contenedores high/low
   - eliminar contenedores que exceden thresholds
*/

// ------------------------------------------------------
// Estructuras relacionadas con lectura del módulo kernel
// ------------------------------------------------------

type ProcProcess struct {
	Pid     int    `json:"pid"`
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	VszKb   uint64 `json:"vsz_kb"`
	RssKb   uint64 `json:"rss_kb"`
	MemPct  string `json:"mem_pct"` // "12.34"
	State   string `json:"state,omitempty"`
}

type DockerInfo struct {
	ContainerID string
	Image       string
	Pid         int
	Name        string
}

type ProcSys struct {
	MemTotalKb uint64        `json:"mem_total_kb"`
	MemFreeKb  uint64        `json:"mem_free_kb"`
	MemUsedKb  uint64        `json:"mem_used_kb"`
	Processes  []ProcProcess `json:"processes"`
}

type ProcCont struct {
	MemTotalKb uint64        `json:"mem_total_kb"`
	MemFreeKb  uint64        `json:"mem_free_kb"`
	MemUsedKb  uint64        `json:"mem_used_kb"`
	Containers []ProcProcess `json:"containers"`
}

// ------------------------------------------------------
// Parámetros del proyecto
// ------------------------------------------------------

var CPU_THRESHOLD float64 = 40.0
var MEM_THRESHOLD float64 = 40.0

var MIN_HIGH_CONTAINERS int = 1
var MIN_LOW_CONTAINERS int = 1

// ------------------------------------------------------
// Obtener información docker pid -> info docker
// ------------------------------------------------------

func GetDockerPidMap() (map[int]DockerInfo, error) {
	out, err := RunCommand("docker", "ps", "-q")
	if err != nil {
		return nil, err
	}

	lines := strings.Fields(out)
	result := make(map[int]DockerInfo)

	for _, id := range lines {
		inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"
		out2, err := RunCommand("docker", "inspect", "--format", inspectFmt, id)
		if err != nil {
			continue
		}

		parts := strings.Fields(strings.TrimSpace(out2))
		if len(parts) < 4 {
			continue
		}

		pid, _ := strconv.Atoi(parts[0])
		cid := parts[1]
		image := parts[2]
		name := parts[3]

		result[pid] = DockerInfo{
			ContainerID: cid,
			Image:       image,
			Pid:         pid,
			Name:        name,
		}
	}
	return result, nil
}

// ------------------------------------------------------
// DECISION LOGIC PRINCIPAL
// ------------------------------------------------------

func DecideAndAct(containers []ProcProcess) {

	// Obtener mapeo PID → DockerInfo
	dmap, err := GetDockerPidMap()
	if err != nil {
		log.Printf("Warning: cannot get docker PID map: %v", err)
	}

	// Agrupar
	type CInfo struct {
		Proc   ProcProcess
		Docker DockerInfo
	}
	var detected []CInfo

	for _, p := range containers {
		if d, ok := dmap[p.Pid]; ok {
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			detected = append(detected, CInfo{
				Proc: p,
				Docker: DockerInfo{
					ContainerID: "",
					Image:       p.Cmdline,
					Pid:         p.Pid,
					Name:        p.Name,
				},
			})
		}
	}

	// Contar high/low
	lowCount := 0
	highCount := 0
	for _, c := range detected {
		img := strings.ToLower(c.Docker.Image)

		if strings.Contains(img, "high") ||
			strings.Contains(img, "cpu") ||
			strings.Contains(img, "mem") {
			highCount++
		} else {
			lowCount++
		}
	}

	// Obtener jiffies
	totalJiffies, _ := ReadTotalJiffies()
	now := time.Now()

	type Decision struct {
		C   CInfo
		Mem float64
		Cpu float64
	}

	var candidates []Decision

	// Calcular CPU y MEM por contenedor
	for _, c := range detected {
		memf, _ := ParseMemPct(c.Proc.MemPct)
		procTime, err := ReadProcPidTime(c.Proc.Pid)
		if err != nil {
			procTime = 0
		}

		cpuPct := CalcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)

		candidates = append(candidates, Decision{
			C:   c,
			Mem: memf,
			Cpu: cpuPct,
		})

		// Guardar en DB
		database.InsertContainerRecord(
			c.Docker.ContainerID,
			c.Proc.Pid,
			c.Docker.Image,
			cpuPct,
			memf,
		)
	}

	// Eliminar contenedores excedidos
	for _, cand := range candidates {

		shouldKill := false
		reason := ""

		if cand.Cpu > CPU_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("CPU %.2f > %.2f", cand.Cpu, CPU_THRESHOLD)
		}

		if cand.Mem > MEM_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("MEM %.2f > %.2f", cand.Mem, MEM_THRESHOLD)
		}

		if !shouldKill {
			continue
		}

		if cand.C.Docker.ContainerID == "" {
			continue
		}

		// No eliminar grafana
		if strings.Contains(strings.ToLower(cand.C.Docker.Image), "grafana") {
			continue
		}

		img := strings.ToLower(cand.C.Docker.Image)
		isHigh := strings.Contains(img, "high") ||
			strings.Contains(img, "cpu") ||
			strings.Contains(img, "mem")

		// Respetar mínimos
		if isHigh && highCount <= MIN_HIGH_CONTAINERS {
			continue
		}
		if !isHigh && lowCount <= MIN_LOW_CONTAINERS {
			continue
		}

		// Eliminar
		log.Printf("Deleting container %s: %s", cand.C.Docker.ContainerID, reason)

		out, err := RunCommand("docker", "rm", "-f", cand.C.Docker.ContainerID)
		if err != nil {
			log.Printf("Failed to delete container %s: %v | out=%s",
				cand.C.Docker.ContainerID, err, out)
		} else {
			database.InsertDeletion(cand.C.Docker.ContainerID, reason)
			if isHigh {
				highCount--
			} else {
				lowCount--
			}
		}
	}
}
