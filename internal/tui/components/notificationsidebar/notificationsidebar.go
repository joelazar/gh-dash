package notificationsidebar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
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

func (m *Model) View() string {
	return lipgloss.NewStyle().Width(m.width).Render("")
}
