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
	Bookmarked bool // Whether this notification has been bookmarked
	Subscribed bool // Whether we're subscribed to this notification thread
}

type Reason string

// Format converts GitHub API notification reason codes to human-readable strings
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
		return "Manual"
	case "security_alert":
		return "Security alert"
	case "your_activity":
		return "Your activity"
	default:
		return string(r)
	}
}

// Implement RowData interface
func (n Notification) GetRepoNameWithOwner() string {
	return n.Repository
}

func (n Notification) GetTitle() string {
	return n.Title
}

func (n Notification) GetNumber() int {
	if id, err := strconv.Atoi(n.ID); err == nil {
		return id
	}
	return 0
}

func (n Notification) GetUrl() string {
	return n.URL
}

func (n Notification) GetUpdatedAt() time.Time {
	return n.UpdatedAt
}
