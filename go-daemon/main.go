package main

import (
	"log"
	"so1-daemon/utils"
)

func main() {

	// Generar las 3 imagenes
	if err := utils.BuildImages(); err != nil {
		log.Printf("Warning: load images failed: %v", err)
	}

	// Generar los 10 contenedores
	if err := utils.BuildContainers(); err != nil {
		log.Printf("Warning: load images failed: %v", err)
	}

}
