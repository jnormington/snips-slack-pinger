package main

import (
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
