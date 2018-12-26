//+build !test

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jnormington/snips-slack-pinger/model"
)

var (
	generateConfig = flag.Bool("generate-config", false, "Output config template")
	config         = flag.String("config", "", "Config file to load")
)

func main() {
	flag.Parse()

	if *generateConfig {
		s, err := model.GenerateConfig()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(s)
		os.Exit(0)
	}

	if *config == "" {
		log.Fatal("missing configuration")
	}

	conf, err := model.LoadConfig(*config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("successfully loaded configuration")

	mc := NewMQTTClient(conf)

	log.Println("attempting to connect")
	go mc.ConnectToMQTTBroker()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

Loop:
	for {
		select {
		case err := <-mc.errCh:
			if err != nil {
				log.Fatal(err)
				break Loop
			}
		case sig := <-sigCh:
			log.Println("exiting... from signal", sig)
			break Loop
		}
	}

	mc.client.Disconnect(0)
}
