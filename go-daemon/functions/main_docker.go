package main_docker

// ------------------ Docker helpers ------------------

// build a map pid -> DockerInfo by inspecting running containers
func getDockerPidMap() (map[int]DockerInfo, error) {
	out, err := runCommand("docker", "ps", "-q")
	if err != nil {
		return nil, err
	}
	lines := strings.Fields(out)
	result := make(map[int]DockerInfo)

	for _, id := range lines {
		// get PID, Image, Name
		inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"
		out2, err := runCommand("docker", "inspect", "--format", inspectFmt, id)
		if err != nil {
			// skip if inspect fails for a container
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
