package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if !ExistsConfig() {
		config := NewConfig()
		if err := config.Write(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Generated dmcm_config.yml. Edit it!")
		return
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	manager := NewManager(config)
	if err := manager.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	manager.Close()
}
