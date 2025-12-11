package functions

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"proyecto1/go-daemon/database"
)

const (
	PROC_CONT = "/proc/continfo_so1_202041390"
	PROC_SYS  = "/proc/sysinfo_so1_202041390"
)

/*
   Este archivo contiene:
   - lectura de archivos /proc generados por los módulos del kernel
   - deserialización
   - invocación de la lógica de decisión
   - ejecución de scripts (cron, grafana, imágenes, módulos)
*/

// ------------------------------------------------------
// UNA iteración del loop
// ------------------------------------------------------

func ProcessOnce() error {

	// 1. Leer sysinfo
	sysRaw, err := ReadProcFile(PROC_SYS)
	if err != nil {
		return fmt.Errorf("read sys proc: %v", err)
	}
	sysRaw = SanitizeJSON(sysRaw)

	var sys ProcSys
	if err := json.Unmarshal(sysRaw, &sys); err != nil {
		return fmt.Errorf("parse sys json: %v", err)
	}

	// Guardar métricas en DB
	database.InsertSysMetrics(sys.MemTotalKb, sys.MemFreeKb, sys.MemUsedKb)

	// 2. Leer continfo
	contRaw, err := ReadProcFile(PROC_CONT)
	if err != nil {
		return fmt.Errorf("read cont proc: %v", err)
	}
	contRaw = SanitizeJSON(contRaw)

	var cont ProcCont
	if err := json.Unmarshal(contRaw, &cont); err != nil {
		return fmt.Errorf("parse cont json: %v", err)
	}

	// 3. Analizar y actuar
	DecideAndAct(cont.Containers)

	return nil
}

// ------------------------------------------------------
// Grafana
// ------------------------------------------------------

func IsGrafanaRunning() bool {
	out, err := RunCommand("docker", "ps", "--filter", "name=grafana_so1", "--format", "{{.Names}}")
	if err != nil {
		return false
	}
	return strings.Contains(out, "grafana_so1")
}

func StartGrafana() error {
	log.Println("Iniciando Grafana...")

	out, err := RunCommand("bash", GRAFANA_START_SCRIPT)
	if err != nil {
		return fmt.Errorf("docker-compose up failed: %v | out: %s", err, out)
	}

	log.Println(out)
	return nil
}

// ------------------------------------------------------
// CRONJOB
// ------------------------------------------------------

func CreateCron() error {
	log.Println("Creating cronjob...")

	out, err := RunCommand("bash", CRON_START_SCRIPT)
	if err != nil {
		return fmt.Errorf("start cron failed: %v | out: %s", err, out)
	}

	log.Printf("Cron started: %s", out)
	return nil
}

func RemoveCron() error {
	log.Println("Removing cronjob...")

	out, err := RunCommand("bash", CRON_STOP_SCRIPT)
	if err != nil {
		return fmt.Errorf("stop cron failed: %v | out: %s", err, out)
	}

	log.Printf("Cron removed: %s", out)
	return nil
}

// ------------------------------------------------------
// Kernel modules
// ------------------------------------------------------

func LoadModules() error {
	log.Println("Loading kernel modules...")

	out, err := RunCommand("bash", LOAD_MODULES_SCRIPT)
	if err != nil {
		return fmt.Errorf("load modules failed: %v | out: %s", err, out)
	}

	log.Printf("Modules load output: %s", out)
	return nil
}

// ------------------------------------------------------
// Build Docker Images
// ------------------------------------------------------

func BuildImages() error {
	log.Println("Generate images of docker...")

	out, err := RunCommand("bash", IMAGES_GENERATE_SCRIPT)
	if err != nil {
		return fmt.Errorf("build images failed: %v | out: %s", err, out)
	}

	log.Printf("Images generated: %s", out)
	return nil
}
