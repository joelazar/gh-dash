package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/issuessection"
	"github.com/dlvhdr/gh-dash/v4/ui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

// ToggleSubscription toggles the subscription of the notification
func (m *Model) ToggleSubscription() tea.Cmd {
	if m.notification == nil {
		log.Debug("ToggleSubscription: no notification selected")
		return nil
	}

	notificationID := m.notification.GetNumber()

	taskId := fmt.Sprintf("toggle_subscription_notification_%d", notificationID)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Toggling subscription for notification %d", notificationID),
		FinishedText: fmt.Sprintf("Toggled subscription for notification %d", notificationID),
		State:        context.TaskStart,
		Error:        nil,
	}

	log.Debug("ToggleSubscription: toggling subscription", "threadID", m.notification.ThreadID, "subscribed", m.notification.Subscribed)

	startCmd := m.ctx.StartTask(task)

	return tea.Batch(startCmd, func() tea.Msg {
		var err error
		if m.notification.Subscribed {
			err = data.UnsubscribeFromNotification(m.notification.ThreadID)
		} else {
			err = data.SubscribeForNotification(m.notification.ThreadID)
		}
		if err != nil {
			log.Debug("ToggleSubscription: failed to toggle", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("ToggleSubscription: successfully toggled, returning action message")

		var isSubscribed bool
		if m.notification.Subscribed {
			isSubscribed = false
		} else {
			isSubscribed = true
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: issuessection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: notificationssection.UpdateNotificationMsg{
				NotificationID: fmt.Sprint(notificationID),
				IsSubscribed:   &isSubscribed,
			},
		}
	})
}
