package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/theme"
)

func TestNotificationUnreadStatusRendering(t *testing.T) {
	ctx := &context.ProgramContext{
		Theme: *theme.DefaultTheme,
	}

	tests := map[string]struct {
		unread       bool
		expectedIcon string
	}{
		"unread notification shows filled circle": {
			unread:       true,
			expectedIcon: "●",
		},
		"read notification shows empty circle": {
			unread:       false,
			expectedIcon: "○",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			notification := &Notification{
				Ctx: ctx,
				Data: &data.Notification{
					ID:         "123",
					Title:      "Test Notification",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					Unread:     tc.unread,
					UpdatedAt:  time.Now(),
					URL:        "https://github.com/owner/repo/issues/123",
					ThreadID:   "thread123",
					Bookmarked: false,
					Subscribed: true,
				},
			}

			unreadIcon := notification.renderUnreadIcon()
			require.Contains(t, unreadIcon, tc.expectedIcon)

			row := notification.ToTableRow()
			unreadColumn := row[4]
			require.Contains(t, unreadColumn, tc.expectedIcon)
		})
	}
}
