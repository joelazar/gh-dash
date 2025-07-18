package notificationsidebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

// MarkAsDone marks the current notification as done
func (m *Model) MarkAsDone() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkAsDone: no notification selected")
		return nil
	}

	log.Debug("MarkAsDone: marking notification as done", "threadID", m.notification.ThreadID)

	return func() tea.Msg {
		err := data.MarkNotificationAsDone(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkAsDone: failed to mark as done", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkAsDone: successfully marked as done, returning action message")
		return NotificationActionMsg{
			Action:   "mark_done",
			ThreadID: m.notification.ThreadID,
		}
	}
}
