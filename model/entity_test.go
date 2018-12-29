package model

import (
	"testing"

	"github.com/bluele/slack"
	"github.com/google/go-cmp/cmp"
)

func TestBuildEntityFromSlackUsers(t *testing.T) {
	users := []*slack.User{
		{Profile: &slack.ProfileInfo{RealName: "Jodie Foster"}},
		{Profile: &slack.ProfileInfo{RealName: "Anthony Hopkins"}},
		{Profile: &slack.ProfileInfo{RealName: "Scott Glenn"}, Deleted: true},
		{Profile: &slack.ProfileInfo{RealName: "Ted Levine"}},
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

	// We don't want username only entries to cause some issue
	// So lets ignore them altogether
	t.Run("when user has no profile", func(t *testing.T) {
		users := []*slack.User{
			{Profile: &slack.ProfileInfo{RealName: "Anthony Hopkins"}},
			{Name: "jodie"},
		}
		want := &Entity{
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
