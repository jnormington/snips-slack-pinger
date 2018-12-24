//+build !test

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	log.Println("Successfully loaded configuration")
	NewMQTTClient(conf)
}
