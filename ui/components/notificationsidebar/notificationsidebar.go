package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

const SectionName = "notification_sidebar"

type Model struct {
	ctx          *context.ProgramContext
	notification *data.Notification
	width        int
	sectionId    int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		ctx:   ctx,
		width: 50,
	}
}

func (m *Model) SetRow(notification *data.Notification) {
	m.notification = notification
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return *m, nil
}

// TODO: decide to improve this or disable completely
func (m *Model) View() string {
	if m.notification == nil {
		return "No notification selected"
	}

	content := fmt.Sprintf("**%s**\n\n", m.notification.Title)
	content += fmt.Sprintf("Repository: %s\n", m.notification.Repository)
	content += fmt.Sprintf("Type: %s\n", m.notification.Type)
	content += fmt.Sprintf("Reason: %s\n", m.notification.Reason.Format())
	content += fmt.Sprintf("Unread: %v\n", m.notification.Unread)
	content += fmt.Sprintf("Updated: %s\n", m.notification.UpdatedAt.Format("2006-01-02 15:04"))
	content += fmt.Sprintf("URL: %s\n\n", m.notification.URL)

	return lipgloss.NewStyle().Width(m.width).Render(content)
}
