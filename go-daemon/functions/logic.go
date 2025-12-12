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
		// check docker map by matching pid
		if d, ok := dmap[p.Pid]; ok {
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			// create placeholder DockerInfo using pid and cmdline as image
			detected = append(detected, CInfo{Proc: p, Docker: var_const.DockerInfo{ContainerID: "", Image: p.Cmdline, Pid: p.Pid, Name: p.Name}})
		}
	}

	// count low/high based on image naming heuristic: image contains "low" or "high"
	lowCount := 0
	highCount := 0
	for _, c := range detected {
		img := strings.ToLower(c.Docker.Image)
		if strings.Contains(img, "high") || strings.Contains(img, "cpu") || strings.Contains(img, "mem") {
			highCount++
		} else {
			lowCount++
		}
	}

	// For each detected compute mem% (parse) and cpu% using /proc
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
		database.InsertContainerRecord(c.Docker.ContainerID, c.Proc.Pid, c.Docker.Image, cpuPct, memf)
	}

	// select to delete: those exceeding thresholds, but respect minima and don't kill grafana
	for _, cand := range candidates {
		shouldKill := false
		reason := ""
		if cand.Cpu > var_const.CPU_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("cpu %.2f > %.2f", cand.Cpu, var_const.CPU_THRESHOLD)
		}
		if cand.Mem > var_const.MEM_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("mem %.2f > %.2f", cand.Mem, var_const.MEM_THRESHOLD)
		}
		// ensure we don't drop below minima
		img := strings.ToLower(cand.C.Docker.Image)
		isHigh := strings.Contains(img, "high") || strings.Contains(img, "cpu") || strings.Contains(img, "mem")
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
			if isHigh {
				if highCount <= var_const.MIN_HIGH_CONTAINERS {
					log.Printf("Would delete %s but would violate MIN_HIGH_CONTAINERS (%d)", cand.C.Docker.ContainerID, var_const.MIN_HIGH_CONTAINERS)
					continue
				}
			} else {
				if lowCount <= var_const.MIN_LOW_CONTAINERS {
					log.Printf("Would delete %s but would violate MIN_LOW_CONTAINERS (%d)", cand.C.Docker.ContainerID, var_const.MIN_LOW_CONTAINERS)
					continue
				}
			}
			// perform deletion
			log.Printf("Deleting container %s due to %s (cpu=%.2f mem=%.2f)", cand.C.Docker.ContainerID, reason, cand.Cpu, cand.Mem)
			out, err := utils.RunCommand("docker", "rm", "-f", cand.C.Docker.ContainerID)
			if err != nil {
				log.Printf("Failed to remove container %s: %v | out: %s", cand.C.Docker.ContainerID, err, out)
			} else {
				database.InsertDeletion(cand.C.Docker.ContainerID, reason)
				// adjust counts
				if isHigh {
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
