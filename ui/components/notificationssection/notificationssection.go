package notificationssection

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/notification"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

const SectionType = "notification"

type Model struct {
	section.BaseModel
	Notifications []data.Notification
	CurrentPage   int
	HasNextPage   bool
}

func NewModel(id int, ctx *context.ProgramContext, cfg config.NotificationsSectionConfig, lastUpdated time.Time, createdAt time.Time) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(ctx),
			Singular:    "Notification",
			Plural:      "Notifications",
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Notifications = []data.Notification{}
	m.CurrentPage = 1
	m.HasNextPage = true
	// Set initial loading state to show "Loading notifications..." instead of empty state
	m.SetIsLoading(true)
	return m
}

type UpdateNotificationMsg struct {
	NotificationID string
	IsRead         *bool
	IsDone         *bool
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.IsSearchFocused() {
			switch {
			case msg.Type == tea.KeyCtrlC, msg.Type == tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return &m, blinkCmd

			case msg.Type == tea.KeyEnter:
				newSearchValue := m.SearchBar.Value()
				m.SearchValue = newSearchValue
				m.SetIsSearching(false)
				m.ResetRows()
				m.SetIsLoading(true) // Show loading state while fetching new results

				// Always fetch from page 1 when search changes
				limit := 50 // fallback default
				if m.Ctx != nil && m.Ctx.Config != nil {
					limit = m.Ctx.Config.Defaults.NotificationsLimit
				}
				if m.Config.Limit != nil {
					limit = *m.Config.Limit
				}
				return &m, tea.Batch(FetchNotificationsWithLimits(m.Ctx, m.Id, limit, m.GetFilters()))
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch {
			case msg.Type == tea.KeyCtrlC, msg.Type == tea.KeyEsc:
				m.SetIsPromptConfirmationShown(false)
				return &m, nil

			case msg.Type == tea.KeyEnter:
				m.SetIsPromptConfirmationShown(false)
				return &m, nil
			}

			break
		}

		switch {
		case key.Matches(msg, keys.Keys.Search):
			m.SetIsSearching(true)
			return &m, nil
		}

	case NotificationsFetchedMsg:
		if msg.SectionId == m.Id {
			log.Debug("PERF: NotificationsFetchedMsg received - sectionId: %d, page: %d, isFirstPage: %v, notifications: %d", msg.SectionId, msg.Page, msg.IsFirstPage, len(msg.Notifications))
			processStart := time.Now()
			
			if msg.IsFirstPage {
					// Replace all notifications for first page
				m.SetRows(msg.Notifications)
				m.CurrentPage = 1
			} else {
					// Append notifications for subsequent pages
				m.Notifications = append(m.Notifications, msg.Notifications...)
				m.CurrentPage = msg.Page
				// Update table rows using the new component
				rowStart := time.Now()
				m.Table.SetRows(m.BuildRows())
			}
			// Calculate the limit used for this fetch to determine if there are more pages
			limit := 50 // fallback default
			if m.Ctx != nil && m.Ctx.Config != nil {
				limit = m.Ctx.Config.Defaults.NotificationsLimit
			}
			if m.Config.Limit != nil {
				limit = *m.Config.Limit
			}

			// Check if we can fetch more pages based on:
			// 1. Whether we got a full page of results
			// 2. Whether we've hit the maximum total limit
			maxLimit := 0
			if m.Ctx != nil && m.Ctx.Config != nil {
				maxLimit = m.Ctx.Config.Defaults.NotificationsMaxLimit
			}
			totalNotifications := len(m.Notifications)
			
			gotFullPage := len(msg.Notifications) == limit
			underMaxLimit := maxLimit <= 0 || totalNotifications < maxLimit
			
			m.HasNextPage = gotFullPage && underMaxLimit
			m.SetIsLoading(false)
		}

	case UpdateNotificationMsg:
		for i, curr := range m.Notifications {
			if curr.ID == msg.NotificationID {
				if msg.IsRead != nil {
					if *msg.IsRead {
						curr.Unread = false
					}
					m.Notifications[i] = curr
				}
				if msg.IsDone != nil {
					if *msg.IsDone {
						// remove done notification from list
						m.Notifications = append(m.Notifications[0:i], (m.Notifications[i+1 : len(m.Notifications)])...)
					}
				}
				m.SetIsLoading(false)
				m.Table.SetRows(m.BuildRows())
				break
			}
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.SearchBar = search
	cmd = tea.Batch(cmd, searchCmd)

	return &m, cmd
}

func (m *Model) SetRows(notifications []data.Notification) {
	log.Debug("PERF: SetRows called - setting %d notifications (replacing %d existing)", len(notifications), len(m.Notifications))
	start := time.Now()
	
	m.Notifications = notifications
	rowStart := time.Now()
	m.Table.SetRows(m.BuildRows())
	log.Debug("PERF: SetRows: BuildRows took %v", time.Since(rowStart))
	
	log.Debug("PERF: SetRows COMPLETE in %v - set %d notifications", time.Since(start), len(notifications))
}

func GetSectionColumns(ctx *context.ProgramContext) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Notifications
	// Currently notifications don't have section-specific configs, so we just use defaults

	repoLayout := dLayout.Repo
	titleLayout := dLayout.Title
	reasonLayout := dLayout.Reason
	typeLayout := dLayout.Type
	updatedAtLayout := dLayout.UpdatedAt

	return []table.Column{
		{
			Title:  "Repository",
			Width:  repoLayout.Width,
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  "Type",
			Width:  typeLayout.Width,
			Hidden: typeLayout.Hidden,
		},
		{
			Title:  "Reason",
			Width:  reasonLayout.Width,
			Hidden: reasonLayout.Hidden,
		},
		{
			Title:  "",
			Width:  utils.IntPtr(3),
			Hidden: utils.BoolPtr(false),
		},
		{
			Title:  "Updated",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
	}
}

// Implement section.Section interface methods
func (m Model) GetId() int                  { return m.Id }
func (m Model) GetType() string             { return SectionType }
func (m Model) GetItemSingularForm() string { return "notification" }
func (m Model) GetItemPluralForm() string   { return "notifications" }

func (m Model) GetCurrRow() data.RowData {
	if len(m.Notifications) == 0 {
		return nil
	}
	currRow := m.Table.GetCurrItem()
	if currRow >= len(m.Notifications) {
		return nil
	}
	return &m.Notifications[currRow]
}

func (m Model) FetchNextPageSectionRows() []tea.Cmd {
	log.Debug("PERF: FetchNextPageSectionRows called - currentPage: %d, hasNextPage: %v, totalNotifications: %d", m.CurrentPage, m.HasNextPage, len(m.Notifications))
	start := time.Now()
	
	// Calculate the limit to use - section-specific limit or default
	limit := 50 // fallback default
	if m.Ctx != nil && m.Ctx.Config != nil {
		limit = m.Ctx.Config.Defaults.NotificationsLimit
	}
	if m.Config.Limit != nil {
		limit = *m.Config.Limit
	}
	log.Debug("PERF: FetchNextPageSectionRows: calculated limit: %d", limit)

	// Check if we've already hit the maximum limit
	maxLimit := 0
	if m.Ctx != nil && m.Ctx.Config != nil {
		maxLimit = m.Ctx.Config.Defaults.NotificationsMaxLimit
	}
	log.Debug("PERF: FetchNextPageSectionRows: maxLimit check - maxLimit: %d, currentCount: %d", maxLimit, len(m.Notifications))
	if maxLimit > 0 && len(m.Notifications) >= maxLimit {
		log.Debug("PERF: FetchNextPageSectionRows: max limit reached, returning nil")
		return nil // Don't fetch more if we've reached the max limit
	}

	// If we have no notifications (after reset), fetch first page
	if len(m.Notifications) == 0 {
		log.Debug("PERF: FetchNextPageSectionRows: no notifications, fetching first page")
		if m.Ctx != nil {
			filters := m.GetFilters()
			log.Debug("PERF: FetchNextPageSectionRows: fetching first page with filters: '%s'", filters)
			cmd := FetchNotificationsWithLimits(m.Ctx, m.Id, limit, filters)
			log.Debug("PERF: FetchNextPageSectionRows: returning first page command in %v", time.Since(start))
			return []tea.Cmd{cmd}
		} else {
			// Fallback for tests
			log.Debug("PERF: FetchNextPageSectionRows: using test fallback")
			return []tea.Cmd{FetchNotifications(m.Id, limit, "")}
		}
	}

	// Otherwise, check if we can fetch next page
	if !m.HasNextPage {
		log.Debug("PERF: FetchNextPageSectionRows: no next page available, returning nil")
		return nil
	}
	// Increment the page for the next fetch
	nextPage := m.CurrentPage + 1
	log.Debug("PERF: FetchNextPageSectionRows: fetching next page %d", nextPage)

	if m.Ctx != nil {
		filters := m.GetFilters()
		log.Debug("PERF: FetchNextPageSectionRows: fetching page %d with filters: '%s'", nextPage, filters)
		cmd := FetchNotificationsPaginatedWithLimits(m.Ctx, m.Id, nextPage, limit, filters)
		log.Debug("PERF: FetchNextPageSectionRows: returning page %d command in %v", nextPage, time.Since(start))
		return []tea.Cmd{cmd}
	} else {
		// Fallback for tests
		log.Debug("PERF: FetchNextPageSectionRows: using test fallback for page %d", nextPage)
		return []tea.Cmd{FetchNotificationsPaginated(m.Id, nextPage, limit, "")}
	}
}

// BuildRows implements the Table interface
func (m Model) BuildRows() []table.Row {
	log.Debug("PERF: BuildRows START - building %d notification rows", len(m.Notifications))
	start := time.Now()
	
	rows := make([]table.Row, len(m.Notifications))
	for i, n := range m.Notifications {
		notificationModel := notification.Notification{
			Ctx:  m.Ctx,
			Data: &n,
		}
		rows[i] = notificationModel.ToTableRow()
	}

	log.Debug("PERF: BuildRows COMPLETE in %v - built %d rows", time.Since(start), len(rows))
	return rows
}

// GetPagerContent implements the Section interface
func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v (fetched %v)",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

// GetTotalCount implements the Section interface
func (m Model) GetTotalCount() int {
	return len(m.Notifications)
}

// NumRows implements the Table interface
func (m Model) NumRows() int {
	return len(m.Notifications)
}

// SetIsLoading implements the Table interface
func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m *Model) ResetRows() {
	m.BaseModel.ResetRows()
	m.Notifications = []data.Notification{} // Clear the notifications data
	m.CurrentPage = 1
	m.HasNextPage = true
}

// GetFilters overrides the base implementation to use the current search bar value
// when the user is actively searching, allowing filter updates while typing
func (m *Model) GetFilters() string {
	// If user is searching and has typed something, temporarily use that value
	if m.IsSearching && m.SearchBar.Value() != "" {
		// Temporarily update SearchValue to use the current search bar value
		originalSearchValue := m.SearchValue
		m.SearchValue = m.SearchBar.Value()
		result := m.BaseModel.GetFilters()
		m.SearchValue = originalSearchValue
		return result
	}
	// Otherwise use the committed SearchValue via the base implementation
	return m.BaseModel.GetFilters()
}

// HasRepoNameInConfiguredFilter overrides the base implementation to also check for
// is:repo(<name>) syntax in addition to repo: prefix
func (m *Model) HasRepoNameInConfiguredFilter() bool {
	// Check configured filters for repo: prefix (base behavior)
	filters := m.Config.Filters
	if utils.HasExplicitRepoFilter(filters) {
		return true
	}

	// Also check current SearchValue for both repo: and is:repo() syntax
	searchValue := m.SearchValue
	if m.IsSearching && m.SearchBar.Value() != "" {
		searchValue = m.SearchBar.Value()
	}

	return utils.HasExplicitRepoFilter(searchValue)
}
