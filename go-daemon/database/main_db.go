package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var (
	db     *sql.DB
	dbLock sync.Mutex

	DB_PATH    = "./data/monitor.db"
	SCHEMA_SQL = "./database/schema.sql"
)

func InitDB() error {
	// Crear carpeta data si no existe
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if err := os.MkdirAll("./data", 0755); err != nil {
			return err
		}
	}

	// Abrir SQLite
	var err error
	db, err = sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return err
	}

	// Leer schema.sql
	schemaBytes, err := ioutil.ReadFile(SCHEMA_SQL)
	if err != nil {
		return err
	}

	// Ejecutar schema
	_, err = db.Exec(string(schemaBytes))
	return err
}

func InsertSysMetrics(total, free, used uint64) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, _ = db.Exec(`INSERT INTO sys_metrics(mem_total_kb, mem_free_kb, mem_used_kb, ts)
                    VALUES(?,?,?,?)`, total, free, used, time.Now().Unix())
}

func InsertContainerRecord(containerID string, pid int, image string, cpuPct, memPct float64) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, _ = db.Exec(`INSERT INTO containers(container_id, pid, image, cpu_pct, mem_pct, ts)
                    VALUES(?,?,?,?,?,?)`,
		containerID, pid, image, cpuPct, memPct, time.Now().Unix())
}

func InsertDeletion(containerID, reason string) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, _ = db.Exec(`INSERT INTO deletions(container_id, reason, ts)
                    VALUES(?,?,?)`,
		containerID, reason, time.Now().Unix())
}
