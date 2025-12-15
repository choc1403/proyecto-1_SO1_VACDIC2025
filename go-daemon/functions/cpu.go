package functions

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"so1-daemon/var_const"
	"strconv"
	"strings"
	"time"
)

// En utils/cgroup.go o donde esté definida ReadCgroupCpuTime

func ReadCgroupCpuTime(containerID string) (uint64, error) {
	// Lista de rutas comunes de cgroup para el uso total de CPU (nanosegundos o microsegundos)
	// PROBABLEMENTE la que te sirva sea la última o la penúltima
	paths := []string{
		// 1. Cgroups V1 estándar (falló en tu log)
		fmt.Sprintf("/sys/fs/cgroup/cpuacct/docker/%s/cpuacct.usage", containerID),
		// 2. Cgroups V1 rootless o variante
		fmt.Sprintf("/sys/fs/cgroup/cpuacct/system.slice/docker-%s.scope/cpuacct.usage", containerID),
		// 3. Cgroups V2 (Docker/Systemd) - ¡Muy común!
		fmt.Sprintf("/sys/fs/cgroup/system.slice/docker-%s.scope/cpu.stat", containerID),
		// 4. Cgroups V2 (Unified) con docker ID completo (Raro pero posible)
		fmt.Sprintf("/sys/fs/cgroup/unified/docker/%s/cpu.stat", containerID),
	}

	var lastErr error
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			lastErr = err
			continue // Intentar la siguiente ruta
		}

		// Si usamos cpu.stat (cgroups v2), necesitamos buscar la línea "usage_usec"
		// o "usage_nsec" dentro del archivo.

		// Asumimos que si llegamos aquí, el archivo existe y es válido.
		s := strings.TrimSpace(string(data))

		// Si el archivo es cpu.stat (cgroups v2), el formato es multi-línea (ej: usage_usec 12345678)
		if strings.Contains(path, "cpu.stat") {
			// Buscamos "usage_usec" o "usage_nsec"
			lines := strings.SplitSeq(s, "\n")
			for line := range lines {
				if strings.HasPrefix(line, "usage_usec") || strings.HasPrefix(line, "usage_nsec") {
					parts := strings.Fields(line) // Dividir por espacio (ej: ["usage_usec", "12345678"])
					if len(parts) == 2 {
						// El valor está en parts[1]
						return parseCgroupValue(parts[1], strings.HasPrefix(line, "usage_usec"))
					}
				}
			}
			// Si no encontramos el campo, es un error de formato.
			lastErr = fmt.Errorf("cpu.stat found but missing usage field")
			continue
		} else {
			// Cgroups V1 (cpuacct.usage) - valor simple en nanosegundos
			return parseCgroupValue(s, false) // V1 es típicamente nanosegundos
		}
	}

	return 0, fmt.Errorf("cgroup CPU usage not found, last error: %w", lastErr)
}

// Función auxiliar para parsear y normalizar
func parseCgroupValue(s string, isMicroseconds bool) (uint64, error) {
	nanoseconds, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing cgroup CPU value: %w", err)
	}

	// Si la lectura estaba en Microsegundos (uso_usec), la escalamos a Nanosegundos
	if isMicroseconds {
		nanoseconds *= 1000
	}
	// Si estaba en Nanosegundos (uso_nsec o V1), no hacemos nada.

	return nanoseconds, nil
}

// ReadTotalJiffies lee el archivo /proc/stat y calcula el total de jiffies
// consumidos por el CPU del sistema.
//
// El valor retornado representa la suma acumulada del tiempo de CPU
// (user, nice, system, idle, iowait, irq, softirq, etc.) desde que el
// sistema fue iniciado. Este valor se utiliza como referencia global
// para calcular el porcentaje de uso de CPU de procesos individuales.
func ReadTotalJiffies() (uint64, error) {

	// Abre el archivo /proc/stat, que contiene estadísticas globales del CPU
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	// Asegura el cierre del archivo al finalizar la función
	defer f.Close()

	// Crea un scanner para leer el archivo línea por línea
	scanner := bufio.NewScanner(f)

	// Lee la primera línea, que corresponde a las estadísticas del CPU
	if !scanner.Scan() {
		return 0, errors.New("empty /proc/stat")
	}

	// Obtiene la línea completa del CPU
	line := scanner.Text()

	// Divide la línea en campos separados por espacios
	// Ejemplo: cpu  4705 150 2253 136239 ...
	fields := strings.Fields(line)

	// Se espera que existan al menos los campos básicos del CPU
	if len(fields) < 8 {
		return 0, fmt.Errorf("inesperado /proc/stat line: %s", line)
	}

	// Variable que acumulará el total de jiffies
	var total uint64 = 0

	// Se ignora el primer campo ("cpu") y se suman todos los valores restantes
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		total += v
	}

	// Retorna el total acumulado de jiffies del sistema
	return total, nil
}

// ReadProcPidTime lee el archivo /proc/[pid]/stat y obtiene el tiempo
// total de CPU consumido por un proceso específico.
//
// El valor retornado corresponde a la suma de:
// - utime: tiempo de CPU en modo usuario
// - stime: tiempo de CPU en modo kernel
//
// Ambos valores están expresados en jiffies y se utilizan para calcular
// el porcentaje de uso de CPU del proceso.
func ReadProcPidTime(pid int) (uint64, error) {

	// Construye la ruta al archivo /proc del proceso
	path := fmt.Sprintf("/proc/%d/stat", pid)

	// Lee el contenido completo del archivo
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	// El archivo /proc/[pid]/stat contiene campos donde el nombre del proceso
	// (comm) puede incluir espacios y está encerrado entre paréntesis.
	//
	// Formato:
	// pid (comm) state ... utime stime ...
	//
	// Para evitar errores al dividir por espacios, se busca el último ')'
	// y se procesa el texto a partir de ese punto.
	s := string(b)
	idx := strings.LastIndex(s, ")")
	if idx == -1 {
		return 0, fmt.Errorf("estadística malformada para pid %d", pid)
	}

	// Se omite ") " y se trabaja únicamente con los campos posteriores
	after := s[idx+2:]
	fields := strings.Fields(after)

	// En el formato original:
	// utime es el campo 14 y stime el campo 15
	// Al haber eliminado "pid (comm)", estos campos se desplazan:
	// utime -> fields[11]
	// stime -> fields[12]
	if len(fields) < 15 {

		// Mecanismo de respaldo: análisis ingenuo por espacios
		parts := strings.Fields(s)
		if len(parts) < 15 {
			return 0, fmt.Errorf("campos de estadística inesperados para pid %d", pid)
		}

		// Extrae utime y stime directamente de los campos originales
		u, _ := strconv.ParseUint(parts[13], 10, 64)
		st, _ := strconv.ParseUint(parts[14], 10, 64)

		// Retorna la suma de tiempos de CPU
		return u + st, nil
	}

	// Extrae utime y stime desde el arreglo ajustado
	u, _ := strconv.ParseUint(fields[11], 10, 64)
	st, _ := strconv.ParseUint(fields[12], 10, 64)

	// Retorna el tiempo total de CPU consumido por el proceso
	return u + st, nil
}

func CalcCpuPercent(pid int, curProcTime, curTotal uint64, curTs time.Time) float64 {
	var_const.PrevSamplesLock.Lock()
	defer var_const.PrevSamplesLock.Unlock()

	prev, ok := var_const.PrevSamples[pid]
	if !ok {
		// ... (Almacenar la primera muestra y devolver 0.0)
		var_const.PrevSamples[pid] = var_const.PidCpuSample{
			TotalProcessTime:   curProcTime,
			TotalSystemJiffies: curTotal,
			Timestamp:          curTs,
		}
		return 0.0
	}

	// --- 1. Calcular diferencias ---
	dProc := float64(curProcTime - prev.TotalProcessTime) // Ahora en NANOSEGUNDOS (del cgroup)

	// dTotalJiffies := float64(curTotal - prev.TotalSystemJiffies)

	// --- 2. Normalizar el tiempo transcurrido (dTotal real) ---
	dTime := curTs.Sub(prev.Timestamp).Seconds() // Tiempo real transcurrido en segundos

	// Si la lectura falló o el tiempo no avanzó.
	if dTime <= 0 {
		return 0.0
	}

	// Guardar la muestra actual para el próximo ciclo
	var_const.PrevSamples[pid] = var_const.PidCpuSample{
		TotalProcessTime:   curProcTime,
		TotalSystemJiffies: curTotal, // Se mantiene por si acaso, pero no se usa en el cálculo
		Timestamp:          curTs,
	}

	// --- 3. Calcular el Uso de CPU (Basado en Nanosegundos y Tiempo Real) ---

	// dProc (Nanosegundos consumidos) se compara con dTime (Segundos reales * 10^9 nanosegundos)
	// dProc / (dTime * 1e9)  <-- Proporción de un solo núcleo (0.0 a 1.0)

	// Se utiliza el factor runtime.NumCPU() para obtener el uso total
	// (igual que docker stats, que puede exceder 100% si es multinúcleo)
	//numCPU := float64(runtime.NumCPU())

	// Fórmula para obtener el uso de CPU Total (similar a docker stats)
	// Docker Stats = (dProc / dTime) / 10^9
	// Usando el factor numCPU de Docker (Uso total vs tiempo total de un núcleo)

	// Si quieres replicar Docker Stats (que puede > 100% en sistemas multi-CPU):
	cpuTotal := (dProc / (dTime * 1e9)) * 100.0 // Esto asume que dProc está en nanosegundos

	return cpuTotal // Retorna el porcentaje total (puede ser 400% si tienes 4 CPUs)
}
