package notificationssection_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

func TestSortByRepoKeyBinding(t *testing.T) {
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'S'},
	}

	matches := key.Matches(keyMsg, keys.NotificationKeys.SortByRepo)
	require.True(t, matches, "The 'S' key should match the SortByRepo binding")

	help := keys.NotificationKeys.SortByRepo.Help()
	require.Equal(t, "S", help.Key, "Key should be 'S'")
	require.Equal(t, "sort by repo", help.Desc, "Description should be 'sort by repo'")
}

func TestSortByRepoInFullHelp(t *testing.T) {
	helpBindings := keys.NotificationFullHelp()

	found := false
	for _, binding := range helpBindings {
		if binding.Help().Key == "S" && binding.Help().Desc == "sort by repo" {
			found = true
			break
		}
	}
	require.True(t, found, "SortByRepo should be included in notification help bindings")
}
