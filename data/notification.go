package data

import (
	"strconv"
	"time"
)

type Notification struct {
	ID         string
	Title      string
	Type       string
	Repository string
	Reason     Reason
	Unread     bool
	UpdatedAt  time.Time
	URL        string
	ThreadID   string
	Subscribed bool // Whether we're subscribed to this notification thread
}

type Reason string

// FormatNotificationReason converts GitHub API notification reason codes to human-readable strings
func (r Reason) Format() string {
	switch r {
	case "review_requested":
		return "Review requested"
	case "mention":
		return "Mentioned"
	case "assign":
		return "Assigned"
	case "author":
		return "Author update"
	case "comment":
		return "New comment"
	case "ci_activity":
		return "CI activity"
	case "push":
		return "New push"
	case "team_mention":
		return "Team mentioned"
	case "state_change":
		return "State changed"
	case "subscribed":
		return "Subscribed"
	case "manual":
		return "Manual subscription"
	case "security_alert":
		return "Security alert"
	case "your_activity":
		return "Your activity"
	default:
		// Return the original reason if we don't have a mapping
		// This ensures we don't lose information for new/unknown reasons
		return string(r)
	}
}

// Implement RowData interface
func (n *Notification) GetRepoNameWithOwner() string {
	return n.Repository
}

func (n *Notification) GetTitle() string {
	return n.Title
}

func (n *Notification) GetNumber() int {
	// Notifications don't have numbers like PRs/Issues, but we need to implement this
	// We can use a hash of the ID or just return 0
	if id, err := strconv.Atoi(n.ID); err == nil {
		return id
	}
	return 0
}

func (n *Notification) GetUrl() string {
	return n.URL
}

func (n *Notification) GetUpdatedAt() time.Time {
	return n.UpdatedAt
}
