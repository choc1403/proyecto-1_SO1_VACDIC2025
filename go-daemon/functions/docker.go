package functions

import (
	"fmt"
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

	lines := strings.Fields(out)
	result := make(map[int]var_const.DockerInfo)

	for _, id := range lines {
		inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"
		out2, err := utils.RunCommand("docker", "inspect", "--format", inspectFmt, id)

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

// ExtractContainerID busca la cadena "-id " en la línea de comandos
// de un containerd-shim y devuelve el Container ID que le sigue.
func ExtractContainerID(cmdline string) string {
	const idPrefix = " -id "

	// 1. Buscar la posición de "-id "
	start := strings.Index(cmdline, idPrefix)
	if start == -1 {
		return "" // No es un shim o el formato ha cambiado
	}

	// Calcular el índice donde empieza el ID (después de " -id ")
	idStart := start + len(idPrefix)

	// Obtener la subcadena que comienza con el ID
	idSubstring := cmdline[idStart:]

	// 2. Buscar el primer espacio después del ID para saber dónde termina
	end := strings.Index(idSubstring, " ")

	if end == -1 {
		// Si no hay más espacios, el ID es hasta el final de la cadena
		return idSubstring
	}

	// Devolver la subcadena del ID
	return idSubstring[:end]
}

// GetDockerInfoByID ejecuta 'docker inspect' en un Container ID específico
// y devuelve la información del contenedor.
func GetDockerInfoByID(id string) (var_const.DockerInfo, error) {
	inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"

	// Ejecutar docker inspect con el ID proporcionado
	out, err := utils.RunCommand("docker", "inspect", "--format", inspectFmt, id)

	if err != nil {
		return var_const.DockerInfo{}, err
	}

	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) < 4 {
		// Formato de salida inesperado
		return var_const.DockerInfo{}, fmt.Errorf("unexpected output from docker inspect: %s", out)
	}

	// Asumimos que partes[0] es el PID
	pid, err := strconv.Atoi(parts[0])
	if err != nil {
		return var_const.DockerInfo{}, fmt.Errorf("invalid PID in docker inspect output: %s", parts[0])
	}

	cid := parts[1]
	image := parts[2]
	name := parts[3]

	dockerInfo := var_const.DockerInfo{
		ContainerID: cid,
		Image:       image,
		Pid:         pid,
		Name:        name,
	}

	return dockerInfo, nil
}
