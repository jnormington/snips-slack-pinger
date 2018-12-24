package model

import (
	"encoding/json"
	"os"
)

// Config holds relevant configuration
// for MQTT client and future config options
type Config struct {
	SnipsConfig SnipsConfig `json:"snips_config"`
	MQTTConfig  MQTTConfig  `json:"mqtt_config"`
}

// SnipsConfig holds snips related
// configration like intent name
type SnipsConfig struct {
	SlackIntent string `json:"slack_intent"`
}

// MQTTConfig contains the configuration
// details for the client to connect
type MQTTConfig struct {
	// Required host
	// Example of host tcp://localhost:1883 or just localhost:1833
	Hosts []string `json:"host"`

	// Optional username authentication
	Username string `json:"username"`
	// Optional password authentication
	Password string `json:"password"`
}

func newDefaultConfig() Config {
	return Config{
		SnipsConfig: SnipsConfig{
			SlackIntent: "username:intent_name",
		},
		MQTTConfig: MQTTConfig{
			Hosts: []string{"localhost:1833"},
		},
	}
}

// GenerateConfig generates a blank
// json encoded config template
func GenerateConfig() (string, error) {
	b, err := json.MarshalIndent(newDefaultConfig(), "", "  ")
	return string(b), err
}

// LoadConfig loads path and returns config
// instance or returns an error
func LoadConfig(path string) (Config, error) {
	var conf Config

	file, err := os.Open(path)
	if err != nil {
		return conf, err
	}

	err = json.NewDecoder(file).Decode(&conf)

	return conf, err
}
