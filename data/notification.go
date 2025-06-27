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
	Reason     string
	Unread     bool
	UpdatedAt  time.Time
	URL        string
	ThreadID   string
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
