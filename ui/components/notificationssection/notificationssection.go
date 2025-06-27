package notificationssection

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

const SectionName = "notifications"

type Model struct {
	section.BaseModel
	Notifications []*data.Notification
}

func NewModel(id int, ctx *context.ProgramContext) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:       id,
			Config:   config.SectionConfig{Title: "Notifications", Filters: ""},
			Type:     SectionName,
			Columns:  GetSectionColumns(),
			Singular: "notification",
			Plural:   "notifications",
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
		},
	)
	m.Notifications = []*data.Notification{}
	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case NotificationsFetchedMsg:
		if msg.SectionId == m.Id {
			m.SetNotifications(msg.Notifications)
			m.IsLoading = false
		}
	}

	return &m, cmd
}

func (m *Model) SetNotifications(notifications []*data.Notification) {
	m.Notifications = notifications
	rows := make([]table.Row, len(notifications))
	for i, n := range notifications {
		unreadIcon := " "
		if n.Unread {
			unreadIcon = "●"
		}
		rows[i] = table.Row{
			unreadIcon,
			n.Repository,
			n.Title,
			n.Reason,
			n.Type,
			n.UpdatedAt.Format("2006-01-02"),
		}
	}
	m.Table.SetRows(rows)
}

func GetSectionColumns() []table.Column {
	return []table.Column{
		{Title: "●", Width: &[]int{4}[0]},
		{Title: "Repository", Width: &[]int{20}[0]},
		{Title: "Title", Width: &[]int{40}[0]},
		{Title: "Reason", Width: &[]int{15}[0]},
		{Title: "Type", Width: &[]int{15}[0]},
		{Title: "Updated", Width: &[]int{15}[0]},
	}
}

// Implement section.Section interface methods
func (m Model) GetId() int { return m.Id }
func (m Model) GetType() string { return SectionName }
func (m Model) GetItemSingularForm() string { return "notification" }
func (m Model) GetItemPluralForm() string { return "notifications" }

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

// BuildRows implements the Table interface
func (m Model) BuildRows() []table.Row {
	rows := make([]table.Row, len(m.Notifications))
	for i, n := range m.Notifications {
		unreadIcon := " "
		if n.Unread {
			unreadIcon = "●"
		}
		rows[i] = table.Row{
			unreadIcon,
			n.Repository,
			n.Title,
			n.Reason,
			n.Type,
			n.UpdatedAt.Format("2006-01-02"),
		}
	}
	return rows
}

// FetchNextPageSectionRows implements the Table interface
func (m Model) FetchNextPageSectionRows() []tea.Cmd {
	return []tea.Cmd{func() tea.Msg {
		notifications, err := data.GetNotifications()
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationsFetchedMsg{
			SectionId:     m.Id,
			Notifications: notifications,
		}
	}}
}

// GetPagerContent implements the Section interface 
func (m Model) GetPagerContent() string {
	totalCount := len(m.Notifications)
	if totalCount == 0 {
		return "No notifications"
	}
	current := m.Table.GetCurrItem() + 1
	return fmt.Sprintf("%d of %d", current, totalCount)
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
