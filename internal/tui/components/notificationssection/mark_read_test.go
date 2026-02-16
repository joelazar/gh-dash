package notificationssection_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
)

func TestUpdateNotificationMsg(t *testing.T) {
	// Test that UpdateNotificationMsg has the expected fields
	msg := notificationssection.UpdateNotificationMsg{
		Id:        "123",
		IsRemoved: true,
	}

	require.Equal(t, "123", msg.Id)
	require.True(t, msg.IsRemoved)
}
