package var_const

import (
	"database/sql"
	"sync"
	"time"
)

const (
	PROC_CONT = "/proc/continfo_so1_202041390"
	PROC_SYS  = "/proc/sysinfo_so1_202041390"

	DB_PATH = "./data/monitor.db"

	DOCKER_COMPOSE_F = "docker-compose.yml"

	// Umbrales
	CPU_THRESHOLD = 15.0 // %
	MEM_THRESHOLD = 10.0 // %
	// Mínimos
	MIN_LOW_CONTAINERS  = 3
	MIN_HIGH_CONTAINERS = 2
)

type ProcProcess struct {
	Pid     int    `json:"pid"`
	Name    string `json:"name"`
	Cmdline string `json:"cmdline"`
	VszKb   uint64 `json:"vsz_kb"`
	RssKb   uint64 `json:"rss_kb"`
	MemPct  string `json:"mem_pct"` // format "X.YY"
	State   string `json:"state,omitempty"`
}

type ProcSys struct {
	MemTotalKb uint64        `json:"mem_total_kb"`
	MemFreeKb  uint64        `json:"mem_free_kb"`
	MemUsedKb  uint64        `json:"mem_used_kb"`
	Processes  []ProcProcess `json:"processes"`
}

type ProcCont struct {
	MemTotalKb uint64        `json:"mem_total_kb"`
	MemFreeKb  uint64        `json:"mem_free_kb"`
	MemUsedKb  uint64        `json:"mem_used_kb"`
	Containers []ProcProcess `json:"containers"`
}

// Docker container helper
type DockerInfo struct {
	ContainerID string
	Image       string
	Pid         int
	Name        string
}

type PidCpuSample struct {
	TotalProcessTime   uint64 // ticks
	TotalSystemJiffies uint64
	Timestamp          time.Time
}

var (
	DB     *sql.DB
	DBLock sync.Mutex

	PrevSamples     = make(map[int]PidCpuSample)
	PrevSamplesLock sync.Mutex
)
