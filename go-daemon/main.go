package main

import (
	"log"
	"so1-daemon/utils"
)

func main() {
	// basic logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Daemon starting...")

	// 1) start grafana
	if err := utils.StartGrafana(); err != nil {
		log.Printf("Warning: starting grafana failed: %v", err)
	} else {
		log.Println("Grafana started.")
	}
	/*

			// Inicializar sqlite
			if err := database.InitDB(); err != nil {
				log.Fatalf("DB init error: %v", err)
			}

			log.Println("DB initialized at", var_const.DB_PATH)
			if _, err := os.Stat(utils.ABSPATH("./data")); os.IsNotExist(err) {
				_ = os.MkdirAll("./data", 0755)
			}

			// Generar los 10 contenedores
			if err := utils.CreateCron(); err != nil {
				log.Printf("Warning: create cron failed: %v", err)
			}

			// Cargar Modulos del Kernel
			if err := utils.LoadModules(); err != nil {
				log.Printf("Warning: load modules failed: %v", err)
			}

			// handle signals for graceful shutdown
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

			// main loop ticker
			ticker := time.NewTicker(20 * time.Second)
			defer ticker.Stop()

			if err := functions.ProcessOnce(); err != nil {
				log.Printf("Initial processOnce error: %v", err)
			}

		loop:
			for {
				select {
				case <-ticker.C:
					if err := functions.ProcessOnce(); err != nil {
						log.Printf("processOnce error: %v", err)
					}
				case <-stop:
					log.Println("Received stop signal, cleaning up...")
					break loop
				}
			}

			// cleanup
			if err := utils.RemoveCron(); err != nil {
				log.Printf("Warning removing cron: %v", err)
			}
			log.Println("Daemon exiting.")
	*/

}
