package notificationsidebar

import (
	"github.com/charmbracelet/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

func (m *Model) MarkRead() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkRead: no notification selected")
		return nil
	}

	log.Debug("MarkRead: marking notification as read", "threadID", m.notification.ThreadID)
	
	return func() tea.Msg {
		err := data.MarkNotificationAsRead(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkRead: failed to mark as read", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkRead: successfully marked as read, returning action message")
		return NotificationActionMsg{
			Action:   "mark_read",
			ThreadID: m.notification.ThreadID,
		}
	}
}

type NotificationActionMsg struct {
	Action   string
	ThreadID string
	URL      string // Optional, used for browser actions
}
