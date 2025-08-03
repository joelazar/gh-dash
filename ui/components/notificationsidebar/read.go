package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

// MarkAsDone marks the current notification as read
func (m *Model) MarkAsRead() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkAsRead: no notification selected")
		return nil
	}

	notificationID := m.notification.GetNumber()

	taskId := fmt.Sprintf("mark_as_read_notification_%d", notificationID)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Marking notification %d to read", notificationID),
		FinishedText: fmt.Sprintf("Marked notification %d as read", notificationID),
		State:        context.TaskStart,
		Error:        nil,
	}

	log.Debug("MarkAsRead: marking notification as read", "threadID", m.notification.ThreadID)

	startCmd := m.ctx.StartTask(task)

	return tea.Batch(startCmd, func() tea.Msg {
		err := data.MarkNotificationAsRead(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkAsRead: failed to mark as read", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkAsRead: successfully marked as read, returning action message")

		trueBool := true

		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: notificationssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: notificationssection.UpdateNotificationMsg{
				NotificationID: fmt.Sprint(notificationID),
				IsRead:         &trueBool,
			},
		}
	})
}
