package notificationssection_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

func TestSearchFilterKeys(t *testing.T) {
	// Test that search key is properly defined
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	}
	matches := key.Matches(keyMsg, keys.Keys.Search)
	require.True(t, matches, "The '/' key should match the Search binding")
}

func TestToggleSmartFilteringKey(t *testing.T) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'t'},
	}
	matches := key.Matches(keyMsg, keys.NotificationKeys.ToggleSmartFiltering)
	require.True(t, matches, "The 't' key should match the ToggleSmartFiltering binding")
}
