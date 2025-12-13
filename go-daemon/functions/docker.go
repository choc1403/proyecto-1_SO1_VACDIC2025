package functions

import (
	"log"
	"so1-daemon/utils"
	"so1-daemon/var_const"
	"strconv"
	"strings"
)

func GetDockerPidMap() (map[int]var_const.DockerInfo, error) {
	out, err := utils.RunCommand("docker", "ps", "-q")

	if err != nil {
		return nil, err
	}

	log.Println("Resultado obtenido: ", out)

	lines := strings.Fields(out)
	result := make(map[int]var_const.DockerInfo)

	for _, id := range lines {
		inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"
		out2, err := utils.RunCommand("docker", "inspect", "--format", inspectFmt, id)

		if err != nil {
			continue
		}
		log.Println("Resultado obtenido: ", out2)
		parts := strings.Fields(strings.TrimSpace(out2))
		if len(parts) < 4 {
			continue
		}
		pid, _ := strconv.Atoi(parts[0])
		cid := parts[1]
		image := parts[2]
		name := parts[3]

		result[pid] = var_const.DockerInfo{
			ContainerID: cid,
			Image:       image,
			Pid:         pid,
			Name:        name,
		}

		log.Println("Resultado: ", result[pid])

	}
	return result, nil

}
