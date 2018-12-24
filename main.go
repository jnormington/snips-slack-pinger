package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jnormington/snips-slack-pinger/model"
)

var generateConfig = flag.Bool("generate-config", false, "Output config template")

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
}
