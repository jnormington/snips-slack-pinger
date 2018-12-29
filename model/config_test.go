package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateConfig(t *testing.T) {
	got, err := GenerateConfig()
	if err != nil {
		t.Fatal(err)
	}

	wc := Config{
		SlackConfig: SlackConfig{
			Token: "1234",
		},
		SnipsConfig: SnipsConfig{
			SlackIntent: "username:intent_name",
			SlotName:    "slack_names",
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

func TestLoadConfig(t *testing.T) {
	writeFile := func(t *testing.T, b []byte) string {
		tmpfile, err := ioutil.TempFile("", "example")
		if err != nil {
			t.Fatal(err)
		}

		if _, err := tmpfile.Write(b); err != nil {
			t.Fatal(err)
		}

		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}

		return tmpfile.Name()
	}

	t.Run("file not existing", func(t *testing.T) {
		_, err := LoadConfig("/tmp/fake_file.json")
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "no such file or directory") {
			t.Fatalf("expected error about missing file but got %q", err)
		}
	})

	t.Run("file not valid json", func(t *testing.T) {
		filename := writeFile(t, []byte(`{"key": value}`))
		defer os.Remove(filename)

		_, err := LoadConfig(filename)
		if err == nil {
			t.Fatal(err)
		}

		want := `invalid character 'v' looking for beginning of value`
		if err.Error() != want {
			t.Fatalf("expected error %q but got %q", want, err)
		}
	})

	t.Run("loaded config successfully", func(t *testing.T) {
		filename := writeFile(t, []byte(`{
			"snips_config": {
				"slack_intent": "intent"
			},
			"mqtt_config": {
				"host": [
				"localhost:1833"
				],
				"username": "bad",
				"password": "pass"
			}
		}`))

		defer os.Remove(filename)

		got, err := LoadConfig(filename)
		if err != nil {
			t.Fatal(err)
		}

		want := Config{
			SnipsConfig: SnipsConfig{
				SlackIntent: "intent",
			},
			MQTTConfig: MQTTConfig{
				Hosts:    []string{"localhost:1833"},
				Username: "bad",
				Password: "pass",
			},
		}

		if !cmp.Equal(want, got) {
			t.Error(cmp.Diff(want, got))
		}
	})
}
