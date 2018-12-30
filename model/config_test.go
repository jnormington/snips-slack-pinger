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
			Token:     "1234",
			Username:  "Standup bot",
			EmojiIcon: ":point_right:",
			Messages: []string{
				"Put your skates on… it’s standup!",
			},
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

func TestConfigValidate(t *testing.T) {
	t.Run("when config all invalid", func(t *testing.T) {
		conf := Config{}

		got := conf.Validate()
		if got == nil {
			t.Fatal("expected error but got none")
		}

		want := "Following error(s) with config:\n" +
			" - slack token required" +
			" - at least one slack message required" +
			" - snips slack intent required" +
			" - snips slot name required"

		if got.Error() != want {
			t.Fatal(cmp.Diff(want, got))
		}
	})

	t.Run("when some of config invalid", func(t *testing.T) {
		conf := Config{
			SlackConfig: SlackConfig{
				Token: "1234",
			},
		}

		got := conf.Validate()
		if got == nil {
			t.Fatal("expected error but got none")
		}

		want := "Following error(s) with config:\n" +
			" - at least one slack message required" +
			" - snips slack intent required" +
			" - snips slot name required"

		if got.Error() != want {
			t.Fatal(cmp.Diff(want, got))
		}
	})

	t.Run("when config all valid", func(t *testing.T) {
		conf := Config{
			SlackConfig: SlackConfig{
				Token: "1234",
				Messages: []string{
					"Put your skates on… it’s standup!",
				},
			},
			SnipsConfig: SnipsConfig{
				SlackIntent: "username:intent_name",
				SlotName:    "slack_names",
			},
		}

		err := conf.Validate()
		if err != nil {
			t.Fatal("expected no error but got", err)
		}
	})
}

func TestSlackConfigIsBlacklisted(t *testing.T) {
	specs := []struct {
		config SlackConfig
		in     string
		want   bool
	}{
		{SlackConfig{Blacklist: []string{}}, "", false},
		{SlackConfig{Blacklist: []string{}}, "ABC", false},
		{SlackConfig{Blacklist: []string{"ABC1234", "DEF2344"}}, "ABC123", false},
		{SlackConfig{Blacklist: []string{"ABC1234", "DEF2344"}}, "ABC1234", true},
		{SlackConfig{Blacklist: []string{"ABC1234", "DEF2344"}}, "DEF2344", true},
	}

	for _, s := range specs {
		if got := s.config.IsBlacklisted(s.in); got != s.want {
			t.Errorf("expected %t but got %t", s.want, got)
		}
	}
}
