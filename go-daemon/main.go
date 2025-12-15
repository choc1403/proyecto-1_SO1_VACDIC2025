package main

import (
	"log"
	"os"
	"os/signal"
	"so1-daemon/database"
	"so1-daemon/functions"
	"so1-daemon/utils"
	"so1-daemon/var_const"
	"syscall"
	"time"
)

func main() {
	// Logs Basicos
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Iniciando Daemon...")

	//Inicializar Grafana
	if err := utils.StartGrafana(); err != nil {
		log.Printf("Advertencia: error al iniciar Grafana: %v", err)
	} else {
		log.Println("Grafana Iniciando.")
	}

	// Inicializar sqlite
	if err := database.InitDB(); err != nil {
		log.Fatalf("Error de Incio DB: %v", err)
	}

	log.Println("Base de datos inicializada en", var_const.DB_PATH)
	if _, err := os.Stat(utils.ABSPATH("./data")); os.IsNotExist(err) {
		_ = os.MkdirAll("./data", 0755)
	}

	// Generar los 10 contenedores
	if err := utils.CreateCron(); err != nil {
		log.Printf("Advertencia: error al crear cron: %v", err)
	}

	// Cargar Modulos del Kernel
	if err := utils.LoadModules(); err != nil {
		log.Printf("Advertencia: error al cargar módulos: %v", err)
	}

	// Gestionar señales para un apagado correcto.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Bucle
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	// 1. PRIMERA MEDICIÓN: Solo guarda los datos base (CPU = 0.0)
	if err := functions.ProcessOnce(); err != nil {
		log.Printf("Proceso inicial. Error: %v", err)
	}

loop:
	for {
		select {
		case <-ticker.C:
			log.Println("Loop tick: ejecutando ProcessOnce()...")
			if err := functions.ProcessOnce(); err != nil {
				log.Printf("Error en ProcessOnce(): %v", err)
			}
		case <-stop:
			log.Println("Señal recibida para detener, limpiando...")
			break loop
		}
	}

	// cleanup
	if err := utils.RemoveCron(); err != nil {
		log.Printf("Advertencia al eliminar cron: %v", err)
	}

	if err := utils.StopContainer(); err != nil {
		log.Printf("Advertencia al eliminar contenedores: %v", err)
	}
	log.Println("Salida de Daemon.")

}
