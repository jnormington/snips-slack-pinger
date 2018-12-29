package model

import (
	"github.com/bluele/slack"
)

// Entity contains operations/data for
// injecting entities via mqtt message
type Entity struct {
	Ops [][]interface{} `json:"operations"`
}

func BuildEntityFromSlackUsers(c SnipsConfig, users []*slack.User) *Entity {
	var entries []string

	if len(users) == 0 {
		return nil
	}

	ei := Entity{Ops: make([][]interface{}, 0)}

	for _, u := range users {
		if u == nil || u.Deleted || u.Profile == nil {
			continue
		}

		entries = append(entries, u.Profile.RealName)
	}

	if len(entries) == 0 {
		return nil
	}

	val := []interface{}{
		"addFromVanilla",
		map[string][]string{
			c.SlotName: entries,
		},
	}

	ei.Ops = append(ei.Ops, val)
	return &ei
}
