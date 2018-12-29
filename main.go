//+build !test

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bluele/slack"
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
	sc := slack.New(conf.SlackConfig.Token)
	go func() {
		// Wait for mqtt client to be connected
		// If its failed the the program will exit
		connected := <-mc.connCh

		if connected {
			updateSlackSlotEntity(sc, mc, conf)

			for range time.Tick(time.Hour * 7) {
				updateSlackSlotEntity(sc, mc, conf)
			}
		}
	}()

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
	close(mc.connCh)
	close(mc.errCh)
	close(sigCh)
}

func updateSlackSlotEntity(sc *slack.Slack, mc mqttClient, conf model.Config) {
	log.Println("preparing new slack names for slot")

	users, err := sc.UsersList()
	if err != nil {
		log.Println("get slack users error:", err)
	}

	res := model.BuildEntityFromSlackUsers(conf.SnipsConfig, users)
	if err := mc.PublishEntity(res); err != nil {
		log.Println("publish entity error:", err)
	}

	log.Println("finished processsing slack names slot")
}
