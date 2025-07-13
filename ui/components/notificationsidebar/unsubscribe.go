package notificationsidebar

import (
	"github.com/charmbracelet/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m *Model) Unsubscribe() tea.Cmd {
	if m.notification == nil {
		log.Debug("Unsubscribe: no notification selected")
		return nil
	}

	log.Debug("Unsubscribe: unsubscribing from notification", "threadID", m.notification.ThreadID)
	
	return func() tea.Msg {
		err := data.UnsubscribeFromNotification(m.notification.ThreadID)
		if err != nil {
			log.Debug("Unsubscribe: failed to unsubscribe", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("Unsubscribe: successfully unsubscribed, returning action message")
		return NotificationActionMsg{
			Action:   "unsubscribe",
			ThreadID: m.notification.ThreadID,
		}
	}
}
