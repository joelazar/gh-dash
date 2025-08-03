package notificationssection

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
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
		log.Debug("PERF: FetchNotifications CMD START - sectionId: %d, limit: %d, query: '%s'", sectionId, limit, searchQuery)
		start := time.Now()

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)
		log.Debug("PERF: FetchNotifications: normalized query from '%s' to '%s'", searchQuery, normalizedQuery)

		notifications, err := data.GetNotificationsPaginated(1, limit, normalizedQuery)
		if err != nil {
			log.Debug("PERF: FetchNotifications CMD FAILED in %v - error: %v", time.Since(start), err)
			return constants.ErrMsg{Err: err}
		}

		log.Debug("PERF: FetchNotifications CMD SUCCESS in %v - returning %d notifications", time.Since(start), len(notifications))
		return NotificationsFetchedMsg{
			SectionId:     sectionId,
			Notifications: notifications,
			IsFirstPage:   true,
		}
	}
}

func FetchNotificationsWithLimits(ctx *context.ProgramContext, sectionId int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		log.Debug("PERF: FetchNotificationsWithLimits CMD START - sectionId: %d, limit: %d, query: '%s'", sectionId, limit, searchQuery)
		start := time.Now()

		// Get the max limits from config
		maxLimit := ctx.Config.Defaults.NotificationsMaxLimit
		maxAgeDays := ctx.Config.Defaults.NotificationsMaxAgeDays
		log.Debug("PERF: FetchNotificationsWithLimits: config limits - maxLimit: %d, maxAgeDays: %d", maxLimit, maxAgeDays)

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)
		log.Debug("PERF: FetchNotificationsWithLimits: normalized query from '%s' to '%s'", searchQuery, normalizedQuery)

		notifications, err := data.GetNotificationsWithLimits(limit, maxLimit, maxAgeDays, normalizedQuery)
		if err != nil {
			log.Debug("PERF: FetchNotificationsWithLimits CMD FAILED in %v - error: %v", time.Since(start), err)
			return constants.ErrMsg{Err: err}
		}

		log.Debug("PERF: FetchNotificationsWithLimits CMD SUCCESS in %v - returning %d notifications", time.Since(start), len(notifications))
		return NotificationsFetchedMsg{
			SectionId:     sectionId,
			Notifications: notifications,
			IsFirstPage:   true,
		}
	}
}

func FetchNotificationsPaginated(sectionId int, page int, limit int, searchQuery string) tea.Cmd {
	return func() tea.Msg {
		log.Debug("PERF: FetchNotificationsPaginated CMD START - sectionId: %d, page: %d, limit: %d, query: '%s'", sectionId, page, limit, searchQuery)
		start := time.Now()

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)
		log.Debug("PERF: FetchNotificationsPaginated: normalized query from '%s' to '%s'", searchQuery, normalizedQuery)

		notifications, err := data.GetNotificationsPaginated(page, limit, normalizedQuery)
		if err != nil {
			log.Debug("PERF: FetchNotificationsPaginated CMD FAILED in %v - error: %v", time.Since(start), err)
			return constants.ErrMsg{Err: err}
		}

		log.Debug("PERF: FetchNotificationsPaginated CMD SUCCESS in %v - page %d, returning %d notifications", time.Since(start), page, len(notifications))
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
		log.Debug("PERF: FetchNotificationsPaginatedWithLimits CMD START - sectionId: %d, page: %d, limit: %d, query: '%s'", sectionId, page, limit, searchQuery)
		start := time.Now()

		// Get the max limits from config
		maxLimit := ctx.Config.Defaults.NotificationsMaxLimit
		maxAgeDays := ctx.Config.Defaults.NotificationsMaxAgeDays
		log.Debug("PERF: FetchNotificationsPaginatedWithLimits: config limits - maxLimit: %d, maxAgeDays: %d", maxLimit, maxAgeDays)

		// Normalize filters to support is:repo(<name>) syntax
		normalizedQuery := utils.NormalizeFilters(searchQuery)
		log.Debug("PERF: FetchNotificationsPaginatedWithLimits: normalized query from '%s' to '%s'", searchQuery, normalizedQuery)

		notifications, err := data.GetNotificationsPaginatedWithLimits(page, limit, maxLimit, maxAgeDays, normalizedQuery)
		if err != nil {
			log.Debug("PERF: FetchNotificationsPaginatedWithLimits CMD FAILED in %v - error: %v", time.Since(start), err)
			return constants.ErrMsg{Err: err}
		}

		log.Debug("PERF: FetchNotificationsPaginatedWithLimits CMD SUCCESS in %v - page %d, returning %d notifications", time.Since(start), page, len(notifications))
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
