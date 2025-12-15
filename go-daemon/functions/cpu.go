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

// Nueva función auxiliar para parsear y normalizar
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

func ReadTotalJiffies() (uint64, error) {
	f, err := os.Open("/proc/stat")

	if err != nil {
		return 0, err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	if !scanner.Scan() {
		return 0, errors.New("empty /proc/stat")
	}

	line := scanner.Text()
	fields := strings.Fields(line)

	if len(fields) < 8 {
		return 0, fmt.Errorf("unexpected /proc/stat line: %s", line)
	}

	var total uint64 = 0

	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		total += v
	}

	return total, nil
}

func ReadProcPidTime(pid int) (uint64, error) {
	path := fmt.Sprintf("/proc/%d/stat", pid)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	// fields may include spaces in cmd; use parsing that respects parentheses
	// The format: pid (comm) R ... utime stime cutime cstime ...
	// We find closing ')' then split after that.
	s := string(b)
	idx := strings.LastIndex(s, ")")
	if idx == -1 {
		return 0, fmt.Errorf("malformed stat for pid %d", pid)
	}
	after := s[idx+2:] // skip ") "
	fields := strings.Fields(after)
	// utime is field 13 and stime 14 counting after comm? We already removed proc & comm, so utime is fields[11]? Safer to count original: utime is 14, stime 15 overall.
	// After we removed first two tokens, utime -> fields[11], stime -> fields[12]
	if len(fields) < 15 {
		// fallback: try naive split
		parts := strings.Fields(s)
		if len(parts) < 15 {
			return 0, fmt.Errorf("unexpected stat fields for pid %d", pid)
		}
		u, _ := strconv.ParseUint(parts[13], 10, 64)
		st, _ := strconv.ParseUint(parts[14], 10, 64)
		return u + st, nil
	}
	u, _ := strconv.ParseUint(fields[11], 10, 64)
	st, _ := strconv.ParseUint(fields[12], 10, 64)
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

	// El dTotal basado en Jiffies ya no es necesario si normalizamos por tiempo real.
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
