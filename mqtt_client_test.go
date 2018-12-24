package main

import (
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jnormington/snips-slack-pinger/model"
)

func TestNewMQTTClient(t *testing.T) {
	conf := model.Config{
		SnipsConfig: model.SnipsConfig{
			SlackIntent: "username:intent_name",
		},
		MQTTConfig: model.MQTTConfig{
			Hosts:    []string{"http://example.com:1883", "tcp://localhost:1833"},
			Username: "bad",
			Password: "pass",
		},
	}

	NewMQTTClient(conf)

	t.Run("validate correct options passed", func(t *testing.T) {
		defer func(fn func(*mqtt.ClientOptions) mqtt.Client) {
			mqttClientFn = fn
		}(mqttClientFn)

		var opts *mqtt.ClientOptions
		mqttClientFn = func(o *mqtt.ClientOptions) mqtt.Client {
			opts = o
			return mqtt.NewClient(o)
		}

		NewMQTTClient(conf)

		if opts == nil {
			t.Fatal("expected opts to be supplied")
		}

		u, p := opts.CredentialsProvider()
		if u != conf.MQTTConfig.Username {
			t.Errorf("expected username %q but got %q", conf.MQTTConfig.Username, u)
		}

		if p != conf.MQTTConfig.Password {
			t.Errorf("expected password %q but got %q", conf.MQTTConfig.Password, p)
		}

		gotLen := len(opts.Servers)
		wantLen := len(conf.MQTTConfig.Hosts)
		if gotLen != wantLen {
			t.Errorf("expected hosts %d but got %d", wantLen, gotLen)
		}

		for i, u := range opts.Servers {
			if u.String() != conf.MQTTConfig.Hosts[i] {
				t.Errorf("expected host %q but got %q", conf.MQTTConfig.Hosts[i], u.String())
			}
		}
	})
}
