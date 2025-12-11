package main

import (
	"log"
	"os"
	"so1-daemon/database"
	"so1-daemon/utils"
	"so1-daemon/var_const"
)

func main() {

	// Inicializar sqlite
	if err := database.InitDB(); err != nil {
		log.Fatalf("DB init error: %v", err)
	}

	log.Println("DB initialized at", var_const.DB_PATH)
	if _, err := os.Stat(utils.ABSPATH("./database/data")); os.IsNotExist(err) {
		_ = os.MkdirAll("./database/data", 0755)
	}

	// Generar los 10 contenedores
	if err := utils.CreateCron(); err != nil {
		log.Printf("Warning: create cron failed: %v", err)
	}

	// Cargar Modulos del Kernel
	if err := utils.LoadModules(); err != nil {
		log.Printf("Warning: load modules failed: %v", err)
	}

}
