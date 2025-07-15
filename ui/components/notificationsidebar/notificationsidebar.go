package notificationsidebar

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

type NotificationActionMsg struct {
	Action   string
	ThreadID string
	URL      string // Optional, used for browser actions
}

// MarkAsDone marks the current notification as read
func (m *Model) MarkAsRead() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkRead: no notification selected")
		return nil
	}

	log.Debug("MarkRead: marking notification as read", "threadID", m.notification.ThreadID)

	return func() tea.Msg {
		err := data.MarkNotificationAsRead(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkRead: failed to mark as read", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkRead: successfully marked as read, returning action message")
		return NotificationActionMsg{
			Action:   "mark_read",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// MarkAsDone marks the current notification as done
func (m *Model) MarkAsDone() tea.Cmd {
	if m.notification == nil {
		log.Debug("MarkAsDone: no notification selected")
		return nil
	}

	log.Debug("MarkAsDone: marking notification as done", "threadID", m.notification.ThreadID)

	return func() tea.Msg {
		err := data.MarkNotificationAsDone(m.notification.ThreadID)
		if err != nil {
			log.Debug("MarkAsDone: failed to mark as done", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("MarkAsDone: successfully marked as done, returning action message")
		return NotificationActionMsg{
			Action:   "mark_done",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// ToggleSubscription toggles the subscription of the notification
func (m *Model) ToggleSubscription() tea.Cmd {
	if m.notification == nil {
		log.Debug("ToggleSubscription: no notification selected")
		return nil
	}

	log.Debug("ToggleSubscription: toggling subscription", "threadID", m.notification.ThreadID, "subscribed", m.notification.Subscribed)

	return func() tea.Msg {
		var err error
		if m.notification.Subscribed {
			err = data.UnsubscribeFromNotification(m.notification.ThreadID)
		} else {
			err = data.SubscribeForNotification(m.notification.ThreadID)
		}
		if err != nil {
			log.Error("ToggleSubscription: failed to toggle", "err", err)
			return constants.ErrMsg{Err: err}
		}
		log.Debug("ToggleSubscription: successfully toggled read status, returning action message")
		return NotificationActionMsg{
			Action:   "toggle_subscription",
			ThreadID: m.notification.ThreadID,
		}
	}
}

// OpenInBrowser opens the notification in the default browser
func (m *Model) OpenInBrowser() tea.Cmd {
	if m.notification == nil {
		log.Debug("OpenInBrowser: no notification selected")
		return nil
	}

	if m.notification.URL == "" {
		log.Debug("OpenInBrowser: notification has no URL", "threadID", m.notification.ThreadID)
		return nil
	}

	log.Debug("OpenInBrowser: opening notification", "threadID", m.notification.ThreadID, "url", m.notification.URL)

	return func() tea.Msg {
		// This will be handled by the main UI to open the browser
		log.Debug("OpenInBrowser: returning action message for browser opening")
		return NotificationActionMsg{
			Action:   "open_browser",
			ThreadID: m.notification.ThreadID,
			URL:      m.notification.URL,
		}
	}
}
