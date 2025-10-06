package notificationssection

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
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

		sections = append(sections, &sectionModel)
		// Use FetchNextPageSectionRows to apply smart filtering like PR and Issues sections
		fetchNotificationsCmds = append(fetchNotificationsCmds, sectionModel.FetchNextPageSectionRows()...)
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

func FetchNotificationsPaginatedWithLimits(ctx *context.ProgramContext, sectionId int, page int, limit int, currentCount int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		// Get the max limits from config
		maxLimit := ctx.Config.Defaults.NotificationsMaxLimit
		maxAgeDays := ctx.Config.Defaults.NotificationsMaxAgeDays

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)

		notifications, err := data.GetNotificationsPaginatedWithCurrentCount(page, limit, maxLimit, maxAgeDays, currentCount, normalizedQuery)
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
