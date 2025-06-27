package notificationsidebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m *Model) Unsubscribe() tea.Cmd {
	if m.notification == nil {
		return nil
	}

	return func() tea.Msg {
		err := data.UnsubscribeFromNotification(m.notification.ThreadID)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationActionMsg{
			Action: "unsubscribe",
			ThreadID: m.notification.ThreadID,
		}
	}
}