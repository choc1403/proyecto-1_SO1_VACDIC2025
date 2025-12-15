package functions

import (
	"fmt"
	"so1-daemon/utils"
	"so1-daemon/var_const"
	"strconv"
	"strings"
)

// GetDockerPidMap obtiene un mapeo entre PID del sistema host y
// la información básica de los contenedores Docker en ejecución.
//
// La función ejecuta comandos del CLI de Docker para:
// 1) Obtener los IDs de los contenedores activos
// 2) Consultar el PID principal de cada contenedor en el host
// 3) Asociar dicho PID con la información del contenedor
//
// Retorna un mapa donde la clave es el PID del proceso del contenedor
// y el valor es una estructura DockerInfo con sus metadatos.
func GetDockerPidMap() (map[int]var_const.DockerInfo, error) {

	// Ejecuta `docker ps -q` para obtener únicamente los IDs
	// de los contenedores actualmente en ejecución
	out, err := utils.RunCommand("docker", "ps", "-q")
	if err != nil {
		return nil, err
	}

	// Divide la salida por espacios/saltos de línea
	// Cada elemento corresponde a un ID de contenedor
	lines := strings.Fields(out)

	// Mapa resultado: PID del contenedor -> información del contenedor
	result := make(map[int]var_const.DockerInfo)

	// Recorre cada ID de contenedor activo
	for _, id := range lines {

		// Formato personalizado para docker inspect:
		// - State.Pid    : PID del proceso principal del contenedor en el host
		// - Id           : ID completo del contenedor
		// - Config.Image : imagen utilizada
		// - Name         : nombre del contenedor
		inspectFmt := "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"

		// Ejecuta docker inspect con el formato definido
		out2, err := utils.RunCommand(
			"docker",
			"inspect",
			"--format",
			inspectFmt,
			id,
		)

		// Si ocurre un error con este contenedor, se omite
		// para no afectar el procesamiento del resto
		if err != nil {
			continue
		}

		// Limpia la salida y la divide en campos individuales
		parts := strings.Fields(strings.TrimSpace(out2))

		// Se esperan al menos 4 campos según el formato definido
		if len(parts) < 4 {
			continue
		}

		// Convierte el PID del contenedor a entero
		pid, _ := strconv.Atoi(parts[0])

		// Extrae los metadatos del contenedor
		cid := parts[1]
		image := parts[2]
		name := parts[3]

		// Asocia el PID del proceso con la información del contenedor
		result[pid] = var_const.DockerInfo{
			ContainerID: cid,
			Image:       image,
			Pid:         pid,
			Name:        name,
		}
	}

	// Retorna el mapa PID -> DockerInfo
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
