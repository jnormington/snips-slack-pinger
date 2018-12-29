package model

import (
	"testing"

	"github.com/bluele/slack"
	"github.com/google/go-cmp/cmp"
)

func TestBuildEntityFromSlackUsers(t *testing.T) {
	users := []*slack.User{
		{Name: "Jodie Foster"},
		{Name: "Anthony Hopkins"},
		{Name: "Scott Glenn", Deleted: true},
		{Name: "Ted Levine"},
	}

	config := SnipsConfig{
		SlotName: "slack_names",
	}

	t.Run("when there are users", func(t *testing.T) {
		want := &Entity{
			Ops: [][]interface{}{
				{
					"addFromVanilla",
					map[string][]string{
						"slack_names": []string{
							"Jodie Foster",
							"Anthony Hopkins",
							"Ted Levine",
						},
					},
				},
			},
		}

		got := BuildEntityFromSlackUsers(config, users)

		if !cmp.Equal(got, want) {
			t.Error(cmp.Diff(want, got))
		}
	})

	t.Run("when all are deleted users or nil", func(t *testing.T) {
		users := []*slack.User{
			{Name: "Jodie Foster", Deleted: true},
			{Name: "Anthony Hopkins", Deleted: true},
			nil,
		}

		var want *Entity
		got := BuildEntityFromSlackUsers(config, users)

		if !cmp.Equal(got, want) {
			t.Error(cmp.Diff(want, got))
		}
	})

	t.Run("no users", func(t *testing.T) {
		var want *Entity
		got := BuildEntityFromSlackUsers(config, []*slack.User{})

		if !cmp.Equal(got, want) {
			t.Error(cmp.Diff(want, got))
		}
	})

	t.Run("slice nil", func(t *testing.T) {
		var want *Entity
		got := BuildEntityFromSlackUsers(config, nil)

		if !cmp.Equal(got, want) {
			t.Error(cmp.Diff(want, got))
		}
	})
}
