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

// MarkAsDone marks the current notification as done
func (m *Model) MarkAsDone() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkAsDone: no notification selected")
		return nil
	}

	log.Debug("MarkAsDone: marking notification as done", "threadID", m.notification.ThreadID)

	notificationID := m.notification.GetNumber()

	taskId := fmt.Sprintf("mark_as_done_notification_%d", notificationID)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Marking notification %d as done", notificationID),
		FinishedText: fmt.Sprintf("Marked notification %d as done", notificationID),
		State:        context.TaskStart,
		Error:        nil,
	}

	startCmd := m.ctx.StartTask(task)

	return tea.Batch(startCmd, func() tea.Msg {
		err := data.MarkNotificationAsDone(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkAsDone: failed to mark as done", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkAsDone: successfully marked as done, returning action message")

		// TODO: better pattern?
		trueBool := true

		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: notificationssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: notificationssection.UpdateNotificationMsg{
				NotificationID: fmt.Sprint(notificationID),
				IsDone:         &trueBool,
			},
		}
	})
}
