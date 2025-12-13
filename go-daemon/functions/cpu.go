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
		// store current and return 0 (no history)
		var_const.PrevSamples[pid] = var_const.PidCpuSample{TotalProcessTime: curProcTime, TotalSystemJiffies: curTotal, Timestamp: curTs}
		return 0.0
	}
	dProc := float64(curProcTime - prev.TotalProcessTime)
	dTotal := float64(curTotal - prev.TotalSystemJiffies)
	// update sample
	var_const.PrevSamples[pid] = var_const.PidCpuSample{TotalProcessTime: curProcTime, TotalSystemJiffies: curTotal, Timestamp: curTs}

	if dTotal <= 0 {

		return 0.0
	}
	// cpu% = (dProc / dTotal) * 100
	return (dProc / dTotal) * 100.0
}
