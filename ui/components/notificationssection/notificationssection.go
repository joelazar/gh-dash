package notificationssection

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

const SectionName = "notifications"

type Model struct {
	section.BaseModel
	Notifications []*data.Notification
	CurrentPage   int
	HasNextPage   bool
}

func NewModel(id int, ctx *context.ProgramContext) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      config.SectionConfig{Title: "Notifications", Filters: ""},
			Type:        SectionName,
			Columns:     GetSectionColumns(ctx),
			Singular:    "notification",
			Plural:      "notifications",
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
		},
	)
	m.Notifications = []*data.Notification{}
	m.CurrentPage = 1
	m.HasNextPage = true
	return m
}

func NewModelWithConfig(id int, ctx *context.ProgramContext, cfg config.NotificationsSectionConfig, lastUpdated time.Time, createdAt time.Time) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionName,
			Columns:     GetSectionColumns(ctx),
			Singular:    "notification",
			Plural:      "notifications",
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Notifications = []*data.Notification{}
	m.CurrentPage = 1
	m.HasNextPage = true
	return m
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
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return &m, tea.Batch(m.FetchNextPageSectionRows()...)
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
			if msg.IsFirstPage {
				// Replace all notifications for first page
				m.SetNotifications(msg.Notifications)
				m.CurrentPage = 1
			} else {
				// Append notifications for subsequent pages
				m.Notifications = append(m.Notifications, msg.Notifications...)
				m.CurrentPage = msg.Page
				// Update table rows
				rows := make([]table.Row, len(m.Notifications))
				for i, n := range m.Notifications {
					unreadIcon := constants.ReadIcon
					if n.Unread {
						unreadIcon = constants.UnreadIcon
					}
					statusIcon := getNotificationStatusIcon(n)
					typeIcon := getNotificationTypeIcon(n.Type)
					rows[i] = table.Row{
						unreadIcon,
						statusIcon,
						n.Repository,
						n.Title,
						utils.FormatNotificationReason(n.Reason),
						typeIcon + " " + n.Type,
						n.UpdatedAt.Format("2006-01-02"),
					}
				}
				m.Table.SetRows(rows)
			}
			// Determine if there are more pages (simple heuristic: if we got a full page, assume there's more)
			m.HasNextPage = len(msg.Notifications) == 50
			m.IsLoading = false
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search
	cmd = tea.Batch(cmd, searchCmd)

	return &m, cmd
}

func (m *Model) SetNotifications(notifications []*data.Notification) {
	m.Notifications = notifications
	rows := make([]table.Row, len(notifications))
	for i, n := range notifications {
		unreadIcon := constants.ReadIcon
		if n.Unread {
			unreadIcon = constants.UnreadIcon
		}
		statusIcon := getNotificationStatusIcon(n)
		typeIcon := getNotificationTypeIcon(n.Type)
		rows[i] = table.Row{
			unreadIcon,
			statusIcon,
			n.Repository,
			n.Title,
			utils.FormatNotificationReason(n.Reason),
			typeIcon + " " + n.Type,
			n.UpdatedAt.Format("2006-01-02"),
		}
	}
	m.Table.SetRows(rows)
}

func GetSectionColumns(ctx *context.ProgramContext) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Notifications
	// Currently notifications don't have section-specific configs, so we just use defaults

	stateLayout := dLayout.State
	repoLayout := dLayout.Repo
	titleLayout := dLayout.Title
	reasonLayout := dLayout.Reason
	typeLayout := dLayout.Type
	updatedAtLayout := dLayout.UpdatedAt

	return []table.Column{
		{
			Title:  "",
			Width:  stateLayout.Width,
			Hidden: stateLayout.Hidden,
		},
		{
			Title:  "",
			Width:  utils.IntPtr(3), // Status column for bookmark/subscription
			Hidden: utils.BoolPtr(false),
		},
		{
			Title:  constants.RepoIcon,
			Width:  repoLayout.Width,
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  constants.ReasonIcon,
			Width:  reasonLayout.Width,
			Hidden: reasonLayout.Hidden,
		},
		{
			Title:  constants.TypeIcon,
			Width:  typeLayout.Width,
			Hidden: typeLayout.Hidden,
		},
		{
			Title:  constants.UpdatedIcon,
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
	}
}

// Implement section.Section interface methods
func (m Model) GetId() int                  { return m.Id }
func (m Model) GetType() string             { return SectionName }
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
	return m.Notifications[currRow]
}

func (m Model) FetchNextPageSectionRows() []tea.Cmd {
	if !m.HasNextPage {
		return nil
	}
	// Increment the page for the next fetch
	nextPage := m.CurrentPage + 1
	
	// Calculate the limit to use - section-specific limit or default
	limit := m.Ctx.Config.Defaults.NotificationsLimit
	if m.Config.Limit != nil {
		limit = *m.Config.Limit
	}
	
	return []tea.Cmd{FetchNotificationsPaginated(m.Id, nextPage, limit, m.GetFilters())}
}

// BuildRows implements the Table interface
func (m Model) BuildRows() []table.Row {
	rows := make([]table.Row, len(m.Notifications))
	for i, n := range m.Notifications {
		unreadIcon := constants.ReadIcon
		if n.Unread {
			unreadIcon = constants.UnreadIcon
		}
		statusIcon := getNotificationStatusIcon(n)
		typeIcon := getNotificationTypeIcon(n.Type)
		rows[i] = table.Row{
			unreadIcon,
			statusIcon,
			n.Repository,
			n.Title,
			utils.FormatNotificationReason(n.Reason),
			typeIcon + " " + n.Type,
			n.UpdatedAt.Format("2006-01-02"),
		}
	}
	return rows
}

// GetPagerContent implements the Section interface
func (m Model) GetPagerContent() string {
	totalCount := len(m.Notifications)
	if totalCount == 0 {
		return fmt.Sprintf("%s No notifications", constants.NotificationIcon)
	}
	current := m.Table.GetCurrItem() + 1
	timeElapsed := utils.TimeElapsed(m.LastUpdated())
	return fmt.Sprintf("%s Updated %s • %s %d/%d • Fetched %d",
		constants.WaitingIcon,
		timeElapsed,
		"notification",
		current,
		totalCount,
		totalCount,
	)
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

// getNotificationTypeIcon returns an appropriate icon for the notification type
func getNotificationTypeIcon(notificationType string) string {
	switch notificationType {
	case "PullRequest":
		return constants.OpenIcon // Use PR open icon
	case "Issue":
		return constants.OpenIcon // Use issue open icon  
	case "Discussion":
		return constants.CommentIcon // Use comment icon for discussions
	case "Release":
		return constants.DonateIcon // Use donate/release icon
	case "RepositoryInvitation":
		return constants.PersonIcon // Use person icon for invitations
	case "SecurityAlert":
		return constants.FailureIcon // Use failure icon for security alerts
	case "CheckSuite", "CheckRun":
		return constants.WaitingIcon // Use waiting icon for CI
	default:
		return constants.NotificationIcon // Default bell icon
	}
}

// getNotificationStatusIcon returns an icon representing bookmark/subscription status
func getNotificationStatusIcon(n *data.Notification) string {
	if n.Bookmarked {
		return constants.BookmarkIcon
	}
	if !n.Subscribed {
		return constants.UnsubscribedIcon
	}
	// If subscribed but not bookmarked, show empty space to keep alignment
	return " "
}
