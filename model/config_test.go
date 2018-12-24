package model

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateConfig(t *testing.T) {
	got, err := GenerateConfig()
	if err != nil {
		t.Fatal(err)
	}

	wc := Config{
		SnipsConfig: SnipsConfig{
			SlackIntent: "username:intent_name",
		},
		MQTTConfig: MQTTConfig{
			Hosts: []string{"localhost:1833"},
		},
	}

	b, err := json.MarshalIndent(wc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	want := string(b)

	if got != want {
		t.Fatalf(cmp.Diff(got, want))
	}
}
