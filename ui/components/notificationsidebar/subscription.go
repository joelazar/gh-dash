package notificationsidebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
)

// ToggleSubscription toggles the subscription of the notification
func (m *Model) ToggleSubscription() tea.Cmd {
	if m.notification == nil {
		log.Debug("ToggleSubscription: no notification selected")
		return nil
	}

	log.Debug("ToggleSubscription: toggling subscription", "threadID", m.notification.ThreadID, "subscribed", m.notification.Subscribed)

	return func() tea.Msg {
		var err error
		if m.notification.Subscribed {
			err = data.UnsubscribeFromNotification(m.notification.ThreadID)
		} else {
			err = data.SubscribeForNotification(m.notification.ThreadID)
		}
		if err != nil {
			log.Error("ToggleSubscription: failed to toggle", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("ToggleSubscription: successfully toggled read status, returning action message")
		return NotificationActionMsg{
			Action:   "toggle_subscription",
			ThreadID: m.notification.ThreadID,
		}
	}
}
