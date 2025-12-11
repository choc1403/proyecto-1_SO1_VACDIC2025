package functions

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
   Estructura para guardar datos previos por PID
   necesarios para calcular %CPU entre muestras
*/
type PidCpuSample struct {
	TotalProcessTime  uint64
	TotalSystemJiffies uint64
	Timestamp         time.Time
}

var (
	prevSamples     = make(map[int]PidCpuSample)
	prevSamplesLock sync.Mutex
)

/* ----------------------------------------------------------
   1. Leer JIFFIES del sistema desde /proc/stat
   ---------------------------------------------------------- */
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

	line := scanner.Text() // primera línea → "cpu  ..."

	fields := strings.Fields(line)
	if len(fields) < 8 {
		return 0, fmt.Errorf("unexpected /proc/stat line: %s", line)
	}

	var total uint64
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		total += v
	}

	return total, nil
}

/* ----------------------------------------------------------
   2. Leer utime + stime desde /proc/<pid>/stat
   ---------------------------------------------------------- */
func ReadProcPidTime(pid int) (uint64, error) {
	path := fmt.Sprintf("/proc/%d/stat", pid)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	s := string(b)

	// encontrar el cierre de ")"
	idx := strings.LastIndex(s, ")")
	if idx == -1 {
		return 0, fmt.Errorf("malformed stat for pid %d", pid)
	}

	after := s[idx+2:] // saltar ") "
	fields := strings.Fields(after)

	// utime → campo 14
	// stime → campo 15
	// después de remover los primeros dos campos, utime=fields[11], stime=fields[12]

	if len(fields) < 15 {
		// fallback
		parts := strings.Fields(s)
		if len(parts) < 15 {
			return 0, fmt.Errorf("unexpected stat format for pid %d", pid)
		}

		u, _ := strconv.ParseUint(parts[13], 10, 64)
		st, _ := strconv.ParseUint(parts[14], 10, 64)
		return u + st, nil
	}

	u, _ := strconv.ParseUint(fields[11], 10, 64)
	st, _ := strconv.ParseUint(fields[12], 10, 64)

	return u + st, nil
}

/* ----------------------------------------------------------
   3. Calcular %CPU desde últimas dos muestras
   ---------------------------------------------------------- */
func CalcCpuPercent(pid int, curProcTime, curTotal uint64, curTs time.Time) float64 {

	prevSamplesLock.Lock()
	defer prevSamplesLock.Unlock()

	prev, ok := prevSamples[pid]
	if !ok {
		// primera muestra → no se puede calcular CPU
		prevSamples[pid] = PidCpuSample{
			TotalProcessTime:  curProcTime,
			TotalSystemJiffies: curTotal,
			Timestamp:         curTs,
		}
		return 0.0
	}

	// diferencias
	dProc := float64(curProcTime - prev.TotalProcessTime)
	dTotal := float64(curTotal - prev.TotalSystemJiffies)

	// almacenar muestra actual
	prevSamples[pid] = PidCpuSample{
		TotalProcessTime:  curProcTime,
		TotalSystemJiffies: curTotal,
		Timestamp:         curTs,
	}

	if dTotal <= 0 {
		return 0.0
	}

	// %CPU = (dProc / dTotal) * 100
	return (dProc / dTotal) * 100.0
}
