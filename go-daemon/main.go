package main

import (
	"log"
	"so1-daemon/utils"
)

func main() {

	if err := utils.TestBash(); err != nil {
		log.Printf("Warning: load modules failed: %v", err)
	}

}
