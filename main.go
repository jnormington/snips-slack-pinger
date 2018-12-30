//+build !test

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bluele/slack"
	"github.com/jnormington/snips-slack-pinger/model"
)

var (
	generateConfig = flag.Bool("generate-config", false, "Output config template")
	config         = flag.String("config", "", "Config file to load")
	dryrun         = flag.Bool("dry-run", false, "Dry run who will be messaged")
	slackUsers     []*slack.User
	slackChannels  []*slack.Channel
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

	if err := conf.Validate(); err != nil {
		log.Fatal(err)
	}

	log.Println("successfully loaded configuration")

	mc := NewMQTTClient(conf, postSlackMessage)
	go updateEntityAndCache(conf, mc)

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

func postSlackMessage(conf model.SlackConfig, name string) error {
	msg := conf.Messages[rand.Intn(len(conf.Messages))]

	var channelID string

	for _, u := range slackUsers {
		if conf.IsBlacklisted(u.Id) {
			continue
		}

		if u != nil &&
			u.Profile != nil &&
			u.Profile.RealName == name {
			channelID = u.Id
			break
		}
	}

	if channelID == "" {
		msg = "@here " + msg
		name = strings.ToLower(name)

		for _, c := range slackChannels {
			if conf.IsBlacklisted(c.Id) {
				continue
			}

			if c != nil && strings.ToLower(c.Name) == name {
				channelID = c.Id
			}
		}
	}

	if channelID == "" {
		return fmt.Errorf("I found no user or channel called %s", name)
	}

	logMsg := fmt.Sprintf("Messaging user/channel %q with ID %q", name, channelID)
	if *dryrun {
		log.Println("[DRYRUN]", logMsg)
		return nil
	}

	log.Println(logMsg)
	sc := slack.New(conf.Token)
	return sc.ChatPostMessage(channelID, msg, &slack.ChatPostMessageOpt{
		LinkNames: "true",
		Username:  conf.Username,
		IconEmoji: conf.EmojiIcon,
	})

	return nil
}

func updateEntityAndCache(conf model.Config, mc mqttClient) {
	var err error
	sc := slack.New(conf.SlackConfig.Token)
	// Wait for mqtt client to be connected
	// If its failed the the program will exit
	connected := <-mc.connCh

	if connected {
		slackUsers, err = sc.UsersList()
		if err != nil {
			log.Println("failed to create slack users cache")
		}

		slackChannels, err = sc.ChannelsList()
		if err != nil {
			log.Println("failed to create slack channels cache")
		}

		updateSlackSlotEntity(mc, slackUsers, conf)
		for range time.Tick(time.Hour * 7) {
			// Update the users/channels cache
			users, err := sc.UsersList()
			if err != nil {
				log.Println("get slack users failed", err)
			} else {
				log.Printf("stored %d users in cache\n", len(users))
				slackUsers = users
			}

			chls, err := sc.ChannelsList()
			if err != nil {
				log.Println("get slack channels failed", err)
			} else {
				slackChannels = chls
			}

			updateSlackSlotEntity(mc, users, conf)
		}
	}
}

func updateSlackSlotEntity(mc mqttClient, users []*slack.User, conf model.Config) {
	log.Println("publishing new slot values")
	res := model.BuildEntityFromSlackUsers(conf.SnipsConfig, users)
	if err := mc.PublishEntity(res); err != nil {
		log.Println("publish entity error:", err)
	}
}
