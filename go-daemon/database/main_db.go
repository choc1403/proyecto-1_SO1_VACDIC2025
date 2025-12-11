package database

import (
	"database/sql"
	"io/ioutil"
	"os"
	"so1-daemon/var_const"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

var SCHEMA_SQL = "./schemas.sql"

func InitDB() error {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if err := os.MkdirAll("./data", 0755); err != nil {
			return err
		}
	}
	var err error
	var_const.DB, err = sql.Open("sqlite3", var_const.DB_PATH)

	if err != nil {
		return err
	}
	// create schema if not exists

	schemaBytes, err := ioutil.ReadFile(SCHEMA_SQL)
	if err != nil {
		return err
	}
	_, err = var_const.DB.Exec(string(schemaBytes))
	return err
}

func InsertSysMetrics(total, free, used uint64) {
	var_const.DBLock.Lock()
	defer var_const.DBLock.Unlock()
	_, _ = var_const.DB.Exec("INSERT INTO sys_metrics(mem_total_kb, mem_free_kb, mem_used_kb, ts) VALUES(?,?,?,?)", total, free, used, time.Now().Unix())
}

func InsertContainerRecord(containerID string, pid int, image string, cpuPct, memPct float64) {
	var_const.DBLock.Lock()
	defer var_const.DBLock.Unlock()
	_, _ = var_const.DB.Exec("INSERT INTO containers(container_id, pid, image, cpu_pct, mem_pct, ts) VALUES(?,?,?,?,?,?)",
		containerID, pid, image, cpuPct, memPct, time.Now().Unix())
}

func InsertDeletion(containerID, reason string) {
	var_const.DBLock.Lock()
	defer var_const.DBLock.Unlock()
	_, _ = var_const.DB.Exec("INSERT INTO deletions(container_id, reason, ts) VALUES(?,?,?)", containerID, reason, time.Now().Unix())
}
