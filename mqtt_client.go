package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jnormington/snips-slack-pinger/model"
)

var mqttClientFn = func(o *mqtt.ClientOptions) mqtt.Client {
	return mqtt.NewClient(o)
}

// NewMQTTClient builds a new mqtt client based
// the on the loaded configuration
func NewMQTTClient(c model.Config) mqtt.Client {
	opts := mqtt.NewClientOptions()

	for _, h := range c.MQTTConfig.Hosts {
		opts.AddBroker(h)
	}

	opts.SetCredentialsProvider(func() (string, string) {
		return c.MQTTConfig.Username, c.MQTTConfig.Password
	})

	return mqttClientFn(opts)
}
