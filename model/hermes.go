package model

// Payload contains the custom payload
// from a mqtt.Message defined by snips
type Payload struct {
	SessionID string                 `json:"sessionId"`
	Values    map[string]interface{} `json:"customData"`
	Intent    Intent                 `json:"intent"`
	Slots     []Slot                 `json:"slots"`
}

// Intent holds name and the
// probability of match made by the NLU
type Intent struct {
	Name        string  `json:"intentName"`
	Probability float64 `json:"probability"`
}

// Slot contains details about extracted
// slots of an intent
type Slot struct {
	Confidence float64        `json:"confidence"`
	Entity     string         `json:"entity"`
	Name       string         `json:"slotName"`
	Range      map[string]int `json:"range"`
	RawValue   string         `json:"raw_value"`
	Value      ValueType      `json:"value"`
}

// ValueType holds a slot value
type ValueType struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// EndSession holds outbound message
// when responding to user and ending session
type EndSession struct {
	SessionID string `json:"sessionId"`
	Text      string `json:"text"`
}
