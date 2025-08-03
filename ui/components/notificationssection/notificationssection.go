package notificationssection

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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

type SortBy int

const (
	SortByUpdated SortBy = iota
	SortByRepo
)

type Model struct {
	section.BaseModel
	Notifications []data.Notification
	CurrentPage   int
	HasNextPage   bool
	SortBy        SortBy
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
	m.SortBy = SortByUpdated
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
				m.SetIsLoading(true)

				limit := 50
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
		case key.Matches(msg, keys.NotificationKeys.SortToggle):
			m.ToggleSort()
			return &m, nil
		}

	case NotificationsFetchedMsg:
		if msg.SectionId == m.Id {
			if msg.IsFirstPage {
				m.SetRows(msg.Notifications)
				m.CurrentPage = 1
			} else {
				// Append new notifications and apply deduplication to the entire set
				allNotifications := append(m.Notifications, msg.Notifications...)
				m.Notifications = data.DeduplicateNotifications(allNotifications)
				m.CurrentPage = msg.Page
				m.Table.SetRows(m.BuildRows())
			}
			limit := 50
			if m.Ctx != nil && m.Ctx.Config != nil {
				limit = m.Ctx.Config.Defaults.NotificationsLimit
			}
			if m.Config.Limit != nil {
				limit = *m.Config.Limit
			}

			maxLimit := 0
			if m.Ctx != nil && m.Ctx.Config != nil {
				maxLimit = m.Ctx.Config.Defaults.NotificationsMaxLimit
			}
			totalNotifications := len(m.Notifications)

			// Check if we got a full page of raw data (before deduplication)
			gotFullPage := len(msg.Notifications) == limit
			
			// Check if we're under the max limit (after deduplication)
			underMaxLimit := maxLimit <= 0 || totalNotifications < maxLimit
			
			// Be more aggressive about fetching when under maxLimit
			// Keep fetching if:
			// 1. We got a full page of raw data, OR
			// 2. We're significantly under maxLimit (by more than one page size)
			//    even if the last fetch wasn't full (could be due to deduplication)
			significantlyUnderLimit := maxLimit > 0 && totalNotifications < (maxLimit - limit)
			
			m.HasNextPage = (gotFullPage || significantlyUnderLimit) && underMaxLimit
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
	// Apply deduplication to keep only the latest notification for each combination
	// of reason, type, repository, and title
	m.Notifications = data.DeduplicateNotifications(notifications)
	m.SortNotifications()
	m.Table.SetRows(m.BuildRows())
}

func GetSectionColumns(ctx *context.ProgramContext) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Notifications

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
	// Calculate the limit to use - section-specific limit or default
	limit := 50 // fallback default
	if m.Ctx != nil && m.Ctx.Config != nil {
		limit = m.Ctx.Config.Defaults.NotificationsLimit
	}
	if m.Config.Limit != nil {
		limit = *m.Config.Limit
	}

	// Check if we've already hit the maximum limit
	maxLimit := 0
	if m.Ctx != nil && m.Ctx.Config != nil {
		maxLimit = m.Ctx.Config.Defaults.NotificationsMaxLimit
	}
	if maxLimit > 0 && len(m.Notifications) >= maxLimit {
		return nil
	}

	// If we have no notifications (after reset), fetch first page
	if len(m.Notifications) == 0 {
		if m.Ctx != nil {
			filters := m.GetFilters()
			cmd := FetchNotificationsWithLimits(m.Ctx, m.Id, limit, filters)
			return []tea.Cmd{cmd}
		} else {
			// Fallback for tests
			return []tea.Cmd{FetchNotifications(m.Id, limit, "")}
		}
	}

	// Otherwise, check if we can fetch next page
	if !m.HasNextPage {
		return nil
	}
	// Increment the page for the next fetch
	nextPage := m.CurrentPage + 1

	if m.Ctx != nil {
		filters := m.GetFilters()
		currentCount := len(m.Notifications) // Pass the current deduplicated count
		cmd := FetchNotificationsPaginatedWithLimits(m.Ctx, m.Id, nextPage, limit, currentCount, filters)
		return []tea.Cmd{cmd}
	} else {
		// Fallback for tests
		return []tea.Cmd{FetchNotificationsPaginated(m.Id, nextPage, limit, "")}
	}
}

func (m *Model) ToggleSort() {
	if m.SortBy == SortByUpdated {
		m.SortBy = SortByRepo
	} else {
		m.SortBy = SortByUpdated
	}
	m.SortNotifications()
	m.Table.SetRows(m.BuildRows())
}

func (m *Model) SortNotifications() {
	switch m.SortBy {
	case SortByRepo:
		// Sort by repository first, then by updated time (newest first)
		slices.SortFunc(m.Notifications, func(a, b data.Notification) int {
			if a.Repository != b.Repository {
				return strings.Compare(a.Repository, b.Repository)
			}
			// If repositories are equal, sort by updated time (newest first)
			return b.UpdatedAt.Compare(a.UpdatedAt)
		})
	case SortByUpdated:
		// Sort by updated time only (newest first)
		slices.SortFunc(m.Notifications, func(a, b data.Notification) int {
			return b.UpdatedAt.Compare(a.UpdatedAt)
		})
	}
}

func (m Model) BuildRows() []table.Row {
	rows := make([]table.Row, len(m.Notifications))
	for i, n := range m.Notifications {
		notificationModel := notification.Notification{
			Ctx:  m.Ctx,
			Data: &n,
		}
		rows[i] = notificationModel.ToTableRow()
	}
	return rows
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	if timeElapsed == "now" {
		timeElapsed = "just now"
	} else {
		timeElapsed = fmt.Sprintf("~%v ago", timeElapsed)
	}
	if m.TotalCount > 0 {
		sortMode := "Updated"
		if m.SortBy == SortByRepo {
			sortMode = "Repo"
		}
		pagerContent = fmt.Sprintf(
			"%v Updated %v • %v %v/%v (fetched %v) • Sort: %v (S)",
			constants.WaitingIcon,
			timeElapsed,
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
			sortMode,
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}

func (m Model) GetTotalCount() int {
	return len(m.Notifications)
}

func (m Model) NumRows() int {
	return len(m.Notifications)
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m *Model) ResetRows() {
	m.BaseModel.ResetRows()
	m.Notifications = []data.Notification{}
	m.CurrentPage = 1
	m.HasNextPage = true
}

func (m *Model) GetFilters() string {
	if m.IsSearching && m.SearchBar.Value() != "" {
		originalSearchValue := m.SearchValue
		m.SearchValue = m.SearchBar.Value()
		result := m.BaseModel.GetFilters()
		m.SearchValue = originalSearchValue
		return result
	}
	return m.BaseModel.GetFilters()
}

func (m *Model) HasRepoNameInConfiguredFilter() bool {
	filters := m.Config.Filters
	if utils.HasExplicitRepoFilter(filters) {
		return true
	}

	searchValue := m.SearchValue
	if m.IsSearching && m.SearchBar.Value() != "" {
		searchValue = m.SearchBar.Value()
	}

	return utils.HasExplicitRepoFilter(searchValue)
}
