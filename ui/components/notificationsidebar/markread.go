package notificationsidebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m *Model) MarkRead() tea.Cmd {
	if m.notification == nil {
		return nil
	}

	return func() tea.Msg {
		err := data.MarkNotificationAsRead(m.notification.ThreadID)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationActionMsg{
			Action: "mark_read",
			ThreadID: m.notification.ThreadID,
		}
	}
}

type NotificationActionMsg struct {
	Action   string
	ThreadID string
	URL      string // Optional, used for browser actions
}