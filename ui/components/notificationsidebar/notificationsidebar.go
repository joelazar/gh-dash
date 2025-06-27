package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

const SectionName = "notification_sidebar"

type Model struct {
	ctx          *context.ProgramContext
	notification *data.Notification
	width        int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		ctx:   ctx,
		width: 50,
	}
}

func (m *Model) SetNotification(notification *data.Notification) {
	m.notification = notification
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return *m, nil
}

func (m *Model) View() string {
	if m.notification == nil {
		return "No notification selected"
	}

	content := fmt.Sprintf("**%s**\n\n", m.notification.Title)
	content += fmt.Sprintf("Repository: %s\n", m.notification.Repository)
	content += fmt.Sprintf("Type: %s\n", m.notification.Type)
	content += fmt.Sprintf("Reason: %s\n", m.notification.Reason)
	content += fmt.Sprintf("Unread: %v\n", m.notification.Unread)
	content += fmt.Sprintf("Updated: %s\n", m.notification.UpdatedAt.Format("2006-01-02 15:04"))
	content += fmt.Sprintf("URL: %s\n\n", m.notification.URL)

	// Add action help
	content += "**Actions:**\n"
	content += "r - Mark as read\n"
	content += "u - Unsubscribe\n"
	content += "b - Bookmark\n"
	content += "d - Mark as done\n"
	content += "t - Toggle read/unread\n"
	content += "o - Open in browser\n"

	return lipgloss.NewStyle().Width(m.width).Render(content)
}

// Bookmark bookmarks the current notification
func (m *Model) Bookmark() tea.Cmd {
	if m.notification == nil {
		return nil
	}

	return func() tea.Msg {
		err := data.BookmarkNotification(m.notification.ThreadID)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationActionMsg{
			Action: "bookmark",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// MarkAsDone marks the current notification as done
func (m *Model) MarkAsDone() tea.Cmd {
	if m.notification == nil {
		return nil
	}

	return func() tea.Msg {
		err := data.MarkNotificationAsDone(m.notification.ThreadID)
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationActionMsg{
			Action: "mark_done",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// ToggleReadStatus toggles the read/unread status of the notification
func (m *Model) ToggleReadStatus() tea.Cmd {
	if m.notification == nil {
		return nil
	}

	return func() tea.Msg {
		var err error
		if m.notification.Unread {
			err = data.MarkNotificationAsRead(m.notification.ThreadID)
		} else {
			err = data.MarkNotificationAsUnread(m.notification.ThreadID)
		}
		if err != nil {
			return constants.ErrMsg{Err: err}
		}
		return NotificationActionMsg{
			Action: "toggle_read",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// OpenInBrowser opens the notification in the default browser
func (m *Model) OpenInBrowser() tea.Cmd {
	if m.notification == nil || m.notification.URL == "" {
		return nil
	}

	return func() tea.Msg {
		// This will be handled by the main UI to open the browser
		return NotificationActionMsg{
			Action: "open_browser",
			ThreadID: m.notification.ThreadID,
			URL: m.notification.URL,
		}
	}
}