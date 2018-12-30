package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jnormington/snips-slack-pinger/model"
)

type slackHandlerFn func(model.SlackConfig, string) error

type mqttClient struct {
	config model.Config
	client mqtt.Client
	errCh  chan error
	connCh chan bool

	slackHandler slackHandlerFn
}

var (
	ErrConnectFail      = errors.New("failed to connect to mqtt broker")
	errInvalidSlotCount = errors.New("missing or too many slots from payload")
	errPublishFailed    = errors.New("failed to publish end session")

	mqttClientFn = func(o *mqtt.ClientOptions) mqtt.Client {
		return mqtt.NewClient(o)
	}

	attempts = 5
)

// NewMQTTClient builds a new mqtt client based
// the on the loaded configuration
func NewMQTTClient(c model.Config, sh slackHandlerFn) mqttClient {
	mqttClt := mqttClient{
		config:       c,
		errCh:        make(chan error),
		connCh:       make(chan bool),
		slackHandler: sh,
	}

	opts := mqtt.NewClientOptions()

	for _, h := range c.MQTTConfig.Hosts {
		opts.AddBroker(h)
	}

	opts.SetCredentialsProvider(func() (string, string) {
		return c.MQTTConfig.Username, c.MQTTConfig.Password
	})

	opts.SetConnectTimeout(5 * time.Second)
	opts.SetOnConnectHandler(mqttClt.ConnectedHandler)
	opts.SetDefaultPublishHandler(mqttClt.MessageHandler)

	mqttClt.client = mqttClientFn(opts)
	return mqttClt
}

// ConnectToMQTTBroker attempts to connect with the broker
// and any connection failure are returned
func (mc mqttClient) ConnectToMQTTBroker() {
	tok := mc.client.Connect()
	if tok.Error() != nil {
		mc.errCh <- tok.Error()
	}

	var try int
	for range time.Tick(time.Second * 1) {
		try++
		if try > attempts {
			mc.errCh <- ErrConnectFail
			mc.connCh <- false
		}

		if mc.client.IsConnected() {
			mc.errCh <- nil
			mc.connCh <- true
			break
		}
	}
}

// ConnectedHandler is called by mqtt.Client when connected
// this handler is responsbile for registering interest in specific
// intents for that it needs to do actions for
func (mc mqttClient) ConnectedHandler(c mqtt.Client) {
	log.Println("connected to MQTT")
	si := mc.config.SnipsConfig.SlackIntent

	log.Printf("registering for events on intent %q\n", si)
	tok := c.Subscribe(fmt.Sprintf("hermes/intent/%s", si), 0, nil)
	if tok.Error() != nil {
		go func(err error) {
			mc.errCh <- err
		}(tok.Error())
	}
}

func (mc mqttClient) MessageHandler(c mqtt.Client, msg mqtt.Message) {
	log.Println("recieved message")
	var p model.Payload

	err := json.Unmarshal(msg.Payload(), &p)
	if err != nil {
		// Don't error just log a handled message failure
		log.Printf("unmarshal message payload error %s\n", err)
		return
	}

	// We won't get here if slot is required
	// but if not set to required we will
	if len(p.Slots) != 1 {
		log.Println(errInvalidSlotCount)
		if err := PublishEndSession(c, p.SessionID, errInvalidSlotCount.Error()); err != nil {
			log.Println(err.Error(), err)
		}
		return
	}

	value := p.Slots[0].Value.Value
	err = mc.slackHandler(mc.config.SlackConfig, value)
	if err != nil {
		PublishEndSession(c, p.SessionID, err.Error())
		return
	}

	if err := PublishEndSession(c, p.SessionID, "I've slacked "+value); err != nil {
		log.Println(err.Error(), err)
	}

	log.Println("processed message")
}

func (mc mqttClient) PublishEntity(e *model.Entity) error {
	b, _ := json.Marshal(e)

	tok := mc.client.Publish("hermes/injection/perform", 0, false, b)

	return tok.Error()
}

func PublishEndSession(c mqtt.Client, sessionID, text string) error {
	end := model.EndSession{
		Text:      text,
		SessionID: sessionID,
	}

	eb, _ := json.Marshal(end)

	ch := "hermes/dialogueManager/endSession"
	tok := c.Publish(ch, 1, false, eb)

	return tok.Error()
}
