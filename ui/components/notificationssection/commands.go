package notificationssection

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func FetchAllSections(ctx *context.ProgramContext) ([]section.Section, tea.Cmd) {
	sectionConfigs := ctx.Config.NotificationsSections
	fetchNotificationsCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections := make([]section.Section, 0, len(sectionConfigs))

	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModelWithConfig(
			i+1, // 0 is reserved for the search section
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		)

		// Calculate the limit to use for API calls
		limit := ctx.Config.Defaults.NotificationsLimit
		if sectionConfig.Limit != nil {
			limit = *sectionConfig.Limit
		}

		sections = append(sections, &sectionModel)
		fetchNotificationsCmds = append(fetchNotificationsCmds, FetchNotifications(i+1, limit, sectionConfig.Filters))
	}

	return sections, tea.Batch(fetchNotificationsCmds...)
}

func FetchNotifications(sectionId int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		notifications, err := data.GetNotificationsPaginated(1, limit, searchQuery)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}

		return NotificationsFetchedMsg{
			SectionId:     sectionId,
			Notifications: notifications,
			IsFirstPage:   true,
		}
	}
}

func FetchNotificationsPaginated(sectionId int, page int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		notifications, err := data.GetNotificationsPaginated(page, limit, searchQuery)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}

		return NotificationsFetchedMsg{
			SectionId:     sectionId,
			Notifications: notifications,
			Page:          page,
			IsFirstPage:   page == 1,
		}
	}
}

type NotificationsFetchedMsg struct {
	SectionId     int
	Notifications []*data.Notification
	Page          int
	IsFirstPage   bool
}
