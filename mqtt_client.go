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

type mqttClient struct {
	config model.Config
	client mqtt.Client
	errCh  chan error
}

var (
	ErrConnectFail = errors.New("failed to connect to mqtt broker")

	mqttClientFn = func(o *mqtt.ClientOptions) mqtt.Client {
		return mqtt.NewClient(o)
	}

	attempts = 5
)

// NewMQTTClient builds a new mqtt client based
// the on the loaded configuration
func NewMQTTClient(c model.Config) mqttClient {
	mqttClt := mqttClient{
		config: c,
		errCh:  make(chan error),
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
		}

		if mc.client.IsConnected() {
			mc.errCh <- nil
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

	log.Printf("registering for events on intent %q", si)
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
		log.Printf("unmarshal message payload error %s", err)
		return
	}

	// Fake slacking user
	time.Sleep(2 * time.Second)
	eb := buildEndSession(p.SessionID, "I slacked. Mimi")

	ch := "hermes/dialogueManager/endSession"
	if t := c.Publish(ch, 1, false, eb); t.Error() != nil {
		log.Printf("failed to publish end session %s", t.Error())
	}

	log.Println("processed message")
}

func buildEndSession(sessionID, text string) []byte {
	end := model.EndSession{
		Text:      text,
		SessionID: sessionID,
	}

	eb, _ := json.Marshal(end)
	return eb
}
