
package main_decision_logic
// ------------------ Decision logic ------------------

// determine containers to stop based on thresholds and keeping minimums
func decideAndAct(containers []ProcProcess) {
	// Build docker pid map
	dmap, err := getDockerPidMap()
	if err != nil {
		log.Printf("Warning: cannot get docker map: %v", err)
	}

	// classify containers discovered (prefer mapping via docker inspect)
	type CInfo struct {
		Proc ProcProcess
		Docker DockerInfo
	}
	var detected []CInfo
	for _, p := range containers {
		// check docker map by matching pid
		if d, ok := dmap[p.Pid]; ok {
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			// create placeholder DockerInfo using pid and cmdline as image
			detected = append(detected, CInfo{Proc: p, Docker: DockerInfo{ContainerID: "", Image: p.Cmdline, Pid: p.Pid, Name: p.Name}})
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
	totalJiffies, _ := readTotalJiffies()
	now := time.Now()
	type decisionCandidate struct {
		C   CInfo
		Mem float64
		Cpu float64
	}
	var candidates []decisionCandidate

	for _, c := range detected {
		memf, _ := parseMemPct(c.Proc.MemPct)
		// compute cpu via /proc/<pid>/stat and /proc/stat
		procTime, err := readProcPidTime(c.Proc.Pid)
		if err != nil {
			// couldn't read stat; set cpu 0
			procTime = 0
		}
		cpuPct := calcCpuPercent(c.Proc.Pid, procTime, totalJiffies, now)
		candidates = append(candidates, decisionCandidate{C: c, Mem: memf, Cpu: cpuPct})

		// Save record in DB
		insertContainerRecord(c.Docker.ContainerID, c.Proc.Pid, c.Docker.Image, cpuPct, memf)
	}

	// select to delete: those exceeding thresholds, but respect minima and don't kill grafana
	for _, cand := range candidates {
		shouldKill := false
		reason := ""
		if cand.Cpu > CPU_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("cpu %.2f > %.2f", cand.Cpu, CPU_THRESHOLD)
		}
		if cand.Mem > MEM_THRESHOLD {
			shouldKill = true
			reason = fmt.Sprintf("mem %.2f > %.2f", cand.Mem, MEM_THRESHOLD)
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
				if highCount <= MIN_HIGH_CONTAINERS {
					log.Printf("Would delete %s but would violate MIN_HIGH_CONTAINERS (%d)", cand.C.Docker.ContainerID, MIN_HIGH_CONTAINERS)
					continue
				}
			} else {
				if lowCount <= MIN_LOW_CONTAINERS {
					log.Printf("Would delete %s but would violate MIN_LOW_CONTAINERS (%d)", cand.C.Docker.ContainerID, MIN_LOW_CONTAINERS)
					continue
				}
			}
			// perform deletion
			log.Printf("Deleting container %s due to %s (cpu=%.2f mem=%.2f)", cand.C.Docker.ContainerID, reason, cand.Cpu, cand.Mem)
			out, err := runCommand("docker", "rm", "-f", cand.C.Docker.ContainerID)
			if err != nil {
				log.Printf("Failed to remove container %s: %v | out: %s", cand.C.Docker.ContainerID, err, out)
			} else {
				insertDeletion(cand.C.Docker.ContainerID, reason)
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
