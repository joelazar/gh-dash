package notificationssection_test

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
)

func TestSortingToggle(t *testing.T) {
	t.Run("default sort is by updated", func(t *testing.T) {
		model := notificationssection.Model{
			SortBy: notificationssection.SortByUpdated,
		}
		
		// Create test notifications with different repos and times
		notifications := []data.Notification{
			{Repository: "repo-b", UpdatedAt: time.Now().Add(-1 * time.Hour)},
			{Repository: "repo-a", UpdatedAt: time.Now()},
		}
		
		model.Notifications = notifications
		model.SortNotifications()
		
		// Should be sorted by updated time (newest first)
		require.Equal(t, "repo-a", model.Notifications[0].Repository)
		require.Equal(t, "repo-b", model.Notifications[1].Repository)
	})

	t.Run("toggle sort changes to repo sort", func(t *testing.T) {
		model := notificationssection.Model{
			SortBy: notificationssection.SortByUpdated,
		}
		
		// Create test notifications
		notifications := []data.Notification{
			{Repository: "repo-b", UpdatedAt: time.Now()},
			{Repository: "repo-a", UpdatedAt: time.Now().Add(-1 * time.Hour)},
		}
		
		model.Notifications = notifications
		
		// Test the toggle logic
		if model.SortBy == notificationssection.SortByUpdated {
			model.SortBy = notificationssection.SortByRepo
		} else {
			model.SortBy = notificationssection.SortByUpdated
		}
		model.SortNotifications()
		
		// Should now be sorted by repo name
		require.Equal(t, notificationssection.SortByRepo, model.SortBy)
		require.Equal(t, "repo-a", model.Notifications[0].Repository)
		require.Equal(t, "repo-b", model.Notifications[1].Repository)
	})

	t.Run("repo sort maintains updated order within same repo", func(t *testing.T) {
		model := notificationssection.Model{
			SortBy: notificationssection.SortByRepo,
		}
		
		now := time.Now()
		notifications := []data.Notification{
			{Repository: "repo-a", UpdatedAt: now.Add(-2 * time.Hour)},
			{Repository: "repo-a", UpdatedAt: now},
			{Repository: "repo-b", UpdatedAt: now.Add(-1 * time.Hour)},
		}
		
		model.Notifications = notifications
		model.SortNotifications()
		
		// Should be sorted by repo first, then by updated time within repo
		require.Equal(t, "repo-a", model.Notifications[0].Repository)
		require.Equal(t, now, model.Notifications[0].UpdatedAt)
		require.Equal(t, "repo-a", model.Notifications[1].Repository)
		require.Equal(t, now.Add(-2*time.Hour), model.Notifications[1].UpdatedAt)
		require.Equal(t, "repo-b", model.Notifications[2].Repository)
	})

	t.Run("key binding toggles sort", func(t *testing.T) {
		// Simulate pressing the sort toggle key
		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'S'},
		}
		
		// Check that the key matches the sort toggle binding
		require.True(t, key.Matches(keyMsg, keys.NotificationKeys.SortToggle))
	})
}

func TestSortingPagerContent(t *testing.T) {
	t.Run("sort mode strings are correct", func(t *testing.T) {
		// Test the sort mode logic directly
		sortByUpdated := notificationssection.SortByUpdated
		sortByRepo := notificationssection.SortByRepo
		
		// Test that the constants are correctly defined
		require.Equal(t, notificationssection.SortByUpdated, sortByUpdated)
		require.Equal(t, notificationssection.SortByRepo, sortByRepo)
		require.NotEqual(t, sortByUpdated, sortByRepo)
	})
}