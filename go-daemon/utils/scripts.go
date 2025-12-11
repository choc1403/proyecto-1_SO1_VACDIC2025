package utils

import (
	"fmt"
	"log"
)

var (
	CRON_START_SCRIPT      = ABSPATH("../../bash/ejecutar_cron.sh")
	CRON_STOP_SCRIPT       = ABSPATH("../../bash/detener_cron.sh")
	LOAD_MODULES_SCRIPT    = ABSPATH("../../bash/cargar_modulos.sh")
	IMAGES_GENERATE_SCRIPT = ABSPATH("../../bash/construir_imagen.sh")
)

func StartGrafana() error {
	// docker-compose up -d
	log.Println("Starting grafana with docker-compose...")
	out, err := RunCommand("docker-compose", "up", "-d")
	if err != nil {
		return fmt.Errorf("docker-compose up failed: %v | out: %s", err, out)
	}
	log.Println("Grafana started.")
	return nil
}

func CreateCron() error {
	// call start_cron script (requires root)

	comando := CRON_START_SCRIPT

	log.Println("Creating cronjob...")
	out, err := RunCommand("bash", comando)
	if err != nil {
		return fmt.Errorf("start cron failed: %v | out: %s", err, out)
	}
	log.Printf("Cron started: %s", out)
	return nil
}

func RemoveCron() error {
	comando := CRON_STOP_SCRIPT

	log.Println("Removing cronjob...")
	out, err := RunCommand("bash", comando)
	if err != nil {
		return fmt.Errorf("stop cron failed: %v | out: %s", err, out)
	}
	log.Printf("Cron removed: %s", out)
	return nil
}

func LoadModules() error {
	comando := LOAD_MODULES_SCRIPT

	log.Println("Loading kernel modules...")
	out, err := RunCommand("bash", comando)
	if err != nil {
		return fmt.Errorf("load modules failed: %v | out: %s", err, out)
	}
	log.Printf("Modules load output: %s", out)
	return nil
}
