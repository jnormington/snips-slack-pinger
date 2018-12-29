package main

import (
	"errors"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/go-cmp/cmp"
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

		if opts.OnConnect == nil {
			t.Fatal("expected connect handler to be set")
		}

		if opts.DefaultPublishHandler == nil {
			t.Fatal("expected default message handler to be set")
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

func TestConnectToMQTTBroker(t *testing.T) {
	defer func(i int) {
		attempts = i
	}(attempts)

	attempts = 1

	t.Run("error from token", func(t *testing.T) {
		wantErr := errors.New("token test error")
		mc := mqttClient{
			client: testMQTTClient{
				token: &testToken{err: wantErr},
			},
			errCh: make(chan error),
		}

		go mc.ConnectToMQTTBroker()

		err := <-mc.errCh
		if err != wantErr {
			t.Fatalf("expected error %q but got %q", wantErr, err)
		}
	})

	t.Run("error when not connected", func(t *testing.T) {
		mc := mqttClient{
			client: testMQTTClient{
				connected: false,
				token:     &testToken{},
			},
			errCh: make(chan error),
		}

		go mc.ConnectToMQTTBroker()

		err := <-mc.errCh
		if err != ErrConnectFail {
			t.Fatalf("expected error %q but got %q", ErrConnectFail, err)
		}
	})

	t.Run("connect success", func(t *testing.T) {
		client := testMQTTClient{connected: true, token: &testToken{}}
		mc := mqttClient{client: client, errCh: make(chan error)}

		go mc.ConnectToMQTTBroker()

		err := <-mc.errCh
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}

		if !client.token.connectCalled {
			t.Fatal("expected connect to be called")
		}
	})
}

func TestConnectedHandler(t *testing.T) {
	t.Run("subscribes to slack intent", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{}}
		mc := buildTestClient(client)

		mc.ConnectedHandler(mc.client)

		want := "hermes/intent/" + mc.config.SnipsConfig.SlackIntent
		if client.token.channel != want {
			t.Fatal(cmp.Diff(want, client.token.channel))
		}
	})

	t.Run("subscribe errors", func(t *testing.T) {
		client := testMQTTClient{
			token: &testToken{
				err: errors.New("subscribe error"),
			},
		}

		mc := buildTestClient(client)

		mc.ConnectedHandler(mc.client)

		got := <-mc.errCh
		if got != client.token.err {
			t.Fatalf("expected error %q but got %q", client.token.err, got)
		}
	})
}

func TestDefaultPublishHandler(t *testing.T) {

	t.Run("json unmarshal errors", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{}}
		mc := buildTestClient(client)

		mc.MessageHandler(mc.client, testMessage{
			payload: []byte(`{"customData": blah}`),
		})

		got := len(client.token.messages)
		if got != 0 {
			t.Fatalf("expected no messages to attempted but got %d", got)
		}
	})

	t.Run("publishes end session", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{}}
		mc := buildTestClient(client)

		mc.MessageHandler(mc.client, testMessage{
			payload: []byte(`{"sessionId": "123", "customData": {}}`),
		})

		got := len(client.token.messages)
		if got != 1 {
			t.Fatalf("expected a message to attempted but got %d", got)
		}

		gotMsg := string(client.token.messages[0].([]byte))
		wantMsg := `{"sessionId":"123","text":"I slacked. Mimi"}`
		if gotMsg != wantMsg {
			t.Fatal(cmp.Diff(wantMsg, gotMsg))
		}

		gotCh := client.token.channel
		wantCh := "hermes/dialogueManager/endSession"
		if gotCh != wantCh {
			t.Fatal(cmp.Diff(wantCh, gotCh))
		}
	})

	t.Run("publishes end session with error", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{err: errors.New("publish message error")}}
		mc := buildTestClient(client)

		mc.MessageHandler(mc.client, testMessage{
			payload: []byte(`{"sessionId": "123", "customData": {}}`),
		})

		got := len(client.token.messages)
		if got != 1 {
			t.Fatalf("expected a message to attempted but got %d", got)
		}

		gotMsg := string(client.token.messages[0].([]byte))
		wantMsg := `{"sessionId":"123","text":"I slacked. Mimi"}`
		if gotMsg != wantMsg {
			t.Fatal(cmp.Diff(wantMsg, gotMsg))
		}

		gotCh := client.token.channel
		wantCh := "hermes/dialogueManager/endSession"
		if gotCh != wantCh {
			t.Fatal(cmp.Diff(wantCh, gotCh))
		}
	})
}

func TestPublishEntity(t *testing.T) {
	entity := &model.Entity{
		Ops: [][]interface{}{
			{
				"addFromVanilla",
				map[string][]string{
					"slack_names": []string{
						"Anthony Hopkins",
					},
				},
			},
		},
	}

	t.Run("publishes entity", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{}}
		mc := buildTestClient(client)
		mc.PublishEntity(entity)

		got := len(client.token.messages)
		if got != 1 {
			t.Fatalf("expected one msg to be published but got %d", got)
		}

		gotMsg := string(client.token.messages[0].([]byte))
		wantMsg := `{"operations":[["addFromVanilla",{"slack_names":["Anthony Hopkins"]}]]}`
		if gotMsg != wantMsg {
			t.Fatal(cmp.Diff(wantMsg, gotMsg))
		}

		gotCh := client.token.channel
		wantCh := "hermes/injection/perform"
		if gotCh != wantCh {
			t.Fatal(cmp.Diff(wantCh, gotCh))
		}
	})

	t.Run("publishes entity", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{err: errors.New("publish error")}}
		mc := buildTestClient(client)
		mc.PublishEntity(entity)

		got := len(client.token.messages)
		if got != 1 {
			t.Fatalf("expected one msg to be published but got %d", got)
		}

		gotMsg := string(client.token.messages[0].([]byte))
		wantMsg := `{"operations":[["addFromVanilla",{"slack_names":["Anthony Hopkins"]}]]}`
		if gotMsg != wantMsg {
			t.Fatal(cmp.Diff(wantMsg, gotMsg))
		}

		gotCh := client.token.channel
		wantCh := "hermes/injection/perform"
		if gotCh != wantCh {
			t.Fatal(cmp.Diff(wantCh, gotCh))
		}
	})
}

func TestPublishEndSession(t *testing.T) {
	t.Run("publishes end session", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{}}
		mc := buildTestClient(client)
		mc.PublishEndSession("1234", "Speak quote")

		got := len(client.token.messages)
		if got != 1 {
			t.Fatalf("expected one msg to be published but got %d", got)
		}

		gotMsg := string(client.token.messages[0].([]byte))
		wantMsg := `{"sessionId":"1234","text":"Speak quote"}`
		if gotMsg != wantMsg {
			t.Fatal(cmp.Diff(wantMsg, gotMsg))
		}

		gotCh := client.token.channel
		wantCh := "hermes/dialogueManager/endSession"
		if gotCh != wantCh {
			t.Fatal(cmp.Diff(wantCh, gotCh))
		}
	})

	t.Run("publishes with error", func(t *testing.T) {
		client := testMQTTClient{token: &testToken{err: errors.New("publish error")}}
		mc := buildTestClient(client)

		err := mc.PublishEndSession("1234", "Some text")
		if err == nil {
			t.Fatal("expected error but got none")
		}
	})
}

func buildTestClient(c testMQTTClient) mqttClient {
	return mqttClient{
		client: c,
		config: model.Config{
			SnipsConfig: model.SnipsConfig{
				SlackIntent: "slack-intent",
			},
		},
		errCh: make(chan error),
	}
}

type testMQTTClient struct {
	connected bool
	token     *testToken
}

type testToken struct {
	err           error
	channel       string
	connectCalled bool
	messages      []interface{}
}

type testMessage struct {
	payload []byte
}

func (ft testToken) Wait() bool                     { return false }
func (ft testToken) Error() error                   { return ft.err }
func (ft testToken) WaitTimeout(time.Duration) bool { return false }

func (f testMQTTClient) IsConnected() bool                    { return f.connected }
func (f testMQTTClient) IsConnectionOpen() bool               { return false }
func (f testMQTTClient) Connect() mqtt.Token                  { f.token.connectCalled = true; return f.token }
func (f testMQTTClient) Disconnect(uint)                      {}
func (f testMQTTClient) Unsubscribe(...string) mqtt.Token     { return f.token }
func (f testMQTTClient) AddRoute(string, mqtt.MessageHandler) {}
func (f testMQTTClient) Publish(ch string, _ byte, _ bool, pl interface{}) mqtt.Token {
	f.token.messages = append(f.token.messages, pl)
	f.token.channel = ch
	return f.token
}
func (f testMQTTClient) Subscribe(c string, _ byte, _ mqtt.MessageHandler) mqtt.Token {
	if f.token.Error() == nil {
		f.token.channel = c
	}
	return f.token
}

func (f testMQTTClient) OptionsReader() mqtt.ClientOptionsReader { return mqtt.ClientOptionsReader{} }
func (f testMQTTClient) SubscribeMultiple(map[string]byte, mqtt.MessageHandler) mqtt.Token {
	return f.token
}

func (tm testMessage) Duplicate() bool   { return false }
func (tm testMessage) Qos() byte         { return 0 }
func (tm testMessage) Retained() bool    { return false }
func (tm testMessage) Topic() string     { return "" }
func (tm testMessage) MessageID() uint16 { return 1 }
func (tm testMessage) Payload() []byte   { return tm.payload }
func (tm testMessage) Ack()              {}
