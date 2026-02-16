package notificationssection_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
)

func TestNotificationRefresh(t *testing.T) {
	// Test that global refresh keys are properly configured for notifications
	require.NotNil(t, keys.Keys.Refresh)
	require.Equal(t, "r", keys.Keys.Refresh.Keys()[0])
	require.Equal(t, "refresh", keys.Keys.Refresh.Help().Desc)

	require.NotNil(t, keys.Keys.RefreshAll)
	require.Equal(t, "R", keys.Keys.RefreshAll.Keys()[0])
	require.Equal(t, "refresh all", keys.Keys.RefreshAll.Help().Desc)

	// Test that mark as read is mapped correctly
	require.NotNil(t, keys.NotificationKeys.MarkAsRead)
	require.Equal(t, "m", keys.NotificationKeys.MarkAsRead.Keys()[0])
	require.Equal(t, "mark as read", keys.NotificationKeys.MarkAsRead.Help().Desc)
}
