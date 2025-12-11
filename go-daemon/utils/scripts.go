package utils

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var (
	CRON_START_SCRIPT         = ABSPATH("../bash/ejecutar_cron.sh")
	CRON_STOP_SCRIPT          = ABSPATH("../bash/detener_cron.sh")
	LOAD_MODULES_SCRIPT       = ABSPATH("../bash/cargar_modulos.sh")
	IMAGES_GENERATE_SCRIPT    = ABSPATH("../bash/construir_imagen.sh")
	GENERATE_CONTAINER_SCRIPT = ABSPATH("../bash/generar_contenedor.sh")
	GRAFANA_COMPOSE_SCRIPT    = ABSPATH("../bash/grafana/generar_grafana.sh")

	TEST = ABSPATH("../bash/prueba.sh")
)

func TestBash() error {
	// call start_cron script (requires root)

	comando := TEST

	log.Println("Realizando Pruebas...")
	out, err := RunCommand("bash", comando)
	if err != nil {
		return fmt.Errorf("start test failed: %v | out: %s", err, out)
	}
	log.Printf("Test started: %s", out)
	return nil
}

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

func BuildImages() error {
	out, err := RunCommand("bash", IMAGES_GENERATE_SCRIPT)
	if err != nil {
		return fmt.Errorf("build images failed: %v | out: %s", err, out)
	}
	log.Printf("Images generated: %s", out)
	return nil
}

func BuildContainers() error {
	out, err := RunCommand("bash", GENERATE_CONTAINER_SCRIPT)
	if err != nil {
		return fmt.Errorf("build container failed: %v | out: %s", err, out)
	}
	log.Printf("Containers generated: %s", out)
	return nil

}

func GenerateGrafanaCompose() error {
	out, err := RunCommand("bash", GRAFANA_COMPOSE_SCRIPT)
	if err != nil {
		return fmt.Errorf("grafana compose generation failed: %v | out: %s", err, out)
	}
	log.Println(out)
	return nil
}

func IsGrafanaRunning() bool {
	out, err := RunCommand("docker", "ps", "--filter", "name=grafana_so1", "--format", "{{.Names}}")
	if err != nil {
		return false
	}
	return strings.Contains(out, "grafana_so1")
}

func PingGrafana() bool {
	conn, err := net.DialTimeout("tcp", "localhost:3000", 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
