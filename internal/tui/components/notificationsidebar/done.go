package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// MarkAsDone marks the current notification as done
func (m *Model) MarkAsDone() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkAsDone: no notification selected")
		return nil
	}

	notificationID := m.notification.ID

	log.Debug("MarkAsDone: marking notification as done", "threadID", m.notification.ThreadID)

	taskId := fmt.Sprintf("mark_as_done_notification_%s", notificationID)
	task := context.Task{
		Id:           taskId,
		StartText:    fmt.Sprintf("Marking notification %s as done", notificationID),
		FinishedText: fmt.Sprintf("Marked notification %s as done", notificationID),
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

		trueBool := true

		return constants.TaskFinishedMsg{
			SectionId:   m.sectionId,
			SectionType: notificationssection.SectionType,
			TaskId:      taskId,
			Err:         err,
			Msg: notificationssection.UpdateNotificationMsg{
				NotificationID: notificationID,
				IsDone:         &trueBool,
			},
		}
	})
}
