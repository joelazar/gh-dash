package notificationssection

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

func FetchAllSections(ctx *context.ProgramContext) ([]section.Section, tea.Cmd) {
	sectionConfigs := ctx.Config.NotificationsSections
	fetchNotificationsCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections := make([]section.Section, 0, len(sectionConfigs))

	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(
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
		// Use the new limits-aware fetching function
		fetchNotificationsCmds = append(fetchNotificationsCmds, FetchNotificationsWithLimits(ctx, i+1, limit, sectionConfig.Filters))
	}

	return sections, tea.Batch(fetchNotificationsCmds...)
}

func FetchNotifications(sectionId int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)

		notifications, err := data.GetNotificationsPaginated(1, limit, normalizedQuery)
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

func FetchNotificationsWithLimits(ctx *context.ProgramContext, sectionId int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		// Get the max limits from config
		maxLimit := ctx.Config.Defaults.NotificationsMaxLimit
		maxAgeDays := ctx.Config.Defaults.NotificationsMaxAgeDays

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)

		notifications, err := data.GetNotificationsWithLimits(limit, maxLimit, maxAgeDays, normalizedQuery)
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
		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)

		notifications, err := data.GetNotificationsPaginated(page, limit, normalizedQuery)
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

func FetchNotificationsPaginatedWithLimits(ctx *context.ProgramContext, sectionId int, page int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		// Get the max limits from config
		maxLimit := ctx.Config.Defaults.NotificationsMaxLimit
		maxAgeDays := ctx.Config.Defaults.NotificationsMaxAgeDays

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)

		notifications, err := data.GetNotificationsPaginatedWithLimits(page, limit, maxLimit, maxAgeDays, normalizedQuery)
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
	Notifications []data.Notification
	Page          int
	IsFirstPage   bool
}
