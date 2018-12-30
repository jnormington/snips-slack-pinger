package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// Config holds relevant configuration
// for MQTT client and future config options
type Config struct {
	SlackConfig SlackConfig `json:"slack_config"`
	SnipsConfig SnipsConfig `json:"snips_config"`
	MQTTConfig  MQTTConfig  `json:"mqtt_config"`
}

// SnipsConfig holds snips related
// configration like intent name
type SnipsConfig struct {
	SlackIntent string `json:"slack_intent"`
	SlotName    string `json:"slot_name"`
}

type SlackConfig struct {
	Token string `json:"token"`

	// Message config options
	Username  string   `json:"username"`
	EmojiIcon string   `json:"emoji_icon"`
	Messages  []string `json:"messages"`

	// Blacklist holds the list of user/channel IDs
	// for which should never be messaged.
	Blacklist []string `json:"blacklist"`
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
			SlotName:    "slack_names",
		},
		SlackConfig: SlackConfig{
			Token:     "1234",
			Username:  "Standup bot",
			EmojiIcon: ":point_right:",
			Messages: []string{
				"Put your skates on… it’s standup!",
			},
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

// Validate validates the config entries but calling
// out to other config parts to fill a buffer of errors
// and returns error if the buf is filled otherwise nil
func (c Config) Validate() error {
	var buf bytes.Buffer

	c.SlackConfig.validate(&buf)
	c.SnipsConfig.validate(&buf)

	if buf.Len() > 0 {
		return fmt.Errorf("Following error(s) with config:\n%s", buf.String())
	}

	return nil
}

func (s SlackConfig) validate(buf *bytes.Buffer) {
	if s.Token == "" {
		buf.WriteString(" - slack token required")
	}

	if len(s.Messages) == 0 {
		buf.WriteString(" - at least one slack message required")
	}
}

func (s SnipsConfig) validate(buf *bytes.Buffer) {
	if s.SlackIntent == "" {
		buf.WriteString(" - snips slack intent required")
	}

	if s.SlotName == "" {
		buf.WriteString(" - snips slot name required")
	}
}

func (s SlackConfig) IsBlacklisted(id string) bool {
	for _, b := range s.Blacklist {
		if id == b {
			return true
		}
	}

	return false
}
