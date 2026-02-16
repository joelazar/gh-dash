package keys

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

func TestSetNotificationSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  NotificationSubjectType
		expected NotificationSubjectType
	}{
		{"none", NotificationSubjectNone, NotificationSubjectNone},
		{"pr", NotificationSubjectPR, NotificationSubjectPR},
		{"issue", NotificationSubjectIssue, NotificationSubjectIssue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetNotificationSubject(tt.subject)
			if notificationSubject != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, notificationSubject)
			}
		})
	}
}

func TestFullHelpIncludesPRKeysForPRSubject(t *testing.T) {
	SetNotificationSubject(NotificationSubjectPR)
	defer SetNotificationSubject(NotificationSubjectNone)

	help := Keys.FullHelp(config.NotificationsView)

	var allKeys []key.Binding
	for _, section := range help {
		allKeys = append(allKeys, section...)
	}

	found := findKeyByHelp(allKeys, "mark as done")
	if !found {
		t.Error("expected notification key 'mark as done' to be present")
	}

	found = findKeyByHelp(allKeys, "diff")
	if !found {
		t.Error("expected PR key 'diff' to be present when viewing PR notification")
	}

	found = findKeyByHelp(allKeys, "approve")
	if !found {
		t.Error("expected PR key 'approve' to be present when viewing PR notification")
	}
}

func TestFullHelpIncludesIssueKeysForIssueSubject(t *testing.T) {
	SetNotificationSubject(NotificationSubjectIssue)
	defer SetNotificationSubject(NotificationSubjectNone)

	help := Keys.FullHelp(config.NotificationsView)

	var allKeys []key.Binding
	for _, section := range help {
		allKeys = append(allKeys, section...)
	}

	found := findKeyByHelp(allKeys, "mark as done")
	if !found {
		t.Error("expected notification key 'mark as done' to be present")
	}

	found = findKeyByHelp(allKeys, "label")
	if !found {
		t.Error("expected Issue key 'label' to be present when viewing Issue notification")
	}

	found = findKeyByHelp(allKeys, "close")
	if !found {
		t.Error("expected Issue key 'close' to be present when viewing Issue notification")
	}
}

func TestFullHelpExcludesPRKeysForNoSubject(t *testing.T) {
	SetNotificationSubject(NotificationSubjectNone)

	help := Keys.FullHelp(config.NotificationsView)

	var allKeys []key.Binding
	for _, section := range help {
		allKeys = append(allKeys, section...)
	}

	found := findKeyByHelp(allKeys, "mark as done")
	if !found {
		t.Error("expected notification key 'mark as done' to be present")
	}

	found = findKeyByHelp(allKeys, "diff")
	if found {
		t.Error("expected PR key 'diff' to NOT be present when no subject is selected")
	}

	found = findKeyByHelp(allKeys, "label")
	if found {
		t.Error("expected Issue key 'label' to NOT be present when no subject is selected")
	}
}

func TestFullHelpForPRViewDoesNotIncludeNotificationKeys(t *testing.T) {
	SetNotificationSubject(NotificationSubjectNone)

	help := Keys.FullHelp(config.PRsView)

	var allKeys []key.Binding
	for _, section := range help {
		allKeys = append(allKeys, section...)
	}

	found := findKeyByHelp(allKeys, "diff")
	if !found {
		t.Error("expected PR key 'diff' to be present in PR view")
	}

	found = findKeyByHelp(allKeys, "toggle bookmark")
	if found {
		t.Error("expected notification key 'toggle bookmark' to NOT be present in PR view")
	}
}

// findKeyByHelp searches for a key binding by its help description
func findKeyByHelp(keys []key.Binding, helpDesc string) bool {
	for _, k := range keys {
		if k.Help().Desc == helpDesc {
			return true
		}
	}
	return false
}
