package notificationssection

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func FetchAllSections(ctx *context.ProgramContext) ([]section.Section, tea.Cmd) {
	return []section.Section{}, func() tea.Msg {
		notifications, err := data.GetNotifications()
		if err != nil {
			return constants.ErrMsg{Err: err}
		}

		return NotificationsFetchedMsg{
			SectionId:     0,
			Notifications: notifications,
		}
	}
}

type NotificationsFetchedMsg struct {
	SectionId     int
	Notifications []*data.Notification
}
