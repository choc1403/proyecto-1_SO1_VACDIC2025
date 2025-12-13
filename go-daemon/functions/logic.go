package functions

import (
	"encoding/json"
	"fmt"
	"log"
	"so1-daemon/database"
	"so1-daemon/utils"
	"so1-daemon/var_const"
	"strings"
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

		log.Println("Que significa p? ", p)
		if d, ok := dmap[p.Pid]; ok {
			detected = append(detected, CInfo{Proc: p, Docker: d})
		} else {
			detected = append(detected, CInfo{
				Proc: p,
				Docker: var_const.DockerInfo{
					ContainerID: "",
					Image:       p.Cmdline,
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
