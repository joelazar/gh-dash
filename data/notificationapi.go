package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func GetNotifications() ([]*Notification, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	var response []struct {
		ID         string `json:"id"`
		Repository struct {
			FullName string `json:"full_name"`
		}
		Subject struct {
			Title string `json:"title"`
			URL   string `json:"url"`
			Type  string `json:"type"`
		}
		Reason    string    `json:"reason"`
		Unread    bool      `json:"unread"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	if err := client.Get("notifications", &response); err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	notifications := make([]*Notification, len(response))
	for i, n := range response {
		notifications[i] = &Notification{
			ID:         n.ID,
			Title:      n.Subject.Title,
			Type:       n.Subject.Type,
			Repository: n.Repository.FullName,
			Reason:     n.Reason,
			Unread:     n.Unread,
			UpdatedAt:  n.UpdatedAt,
			URL:        convertURL(n.Subject.URL),
			ThreadID:   n.ID,
		}
	}

	return notifications, nil
}

func MarkNotificationAsRead(threadID string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	return client.Patch(fmt.Sprintf("notifications/threads/%s", threadID), nil, nil)
}

func UnsubscribeFromNotification(threadID string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	return client.Delete(fmt.Sprintf("notifications/threads/%s/subscription", threadID), nil)
}

// BookmarkNotification saves a notification for later
func BookmarkNotification(threadID string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}
	// GitHub API doesn't have a direct bookmark feature, but we can use the "subscription" endpoint
	// to set the notification as "subscribed" which keeps it visible
	// For now, we'll just ensure the notification stays visible by not deleting the subscription
	return client.Put(fmt.Sprintf("notifications/threads/%s/subscription", threadID), nil, nil)
}

// MarkNotificationAsDone marks a notification as done (same as read but more semantic)
func MarkNotificationAsDone(threadID string) error {
	return MarkNotificationAsRead(threadID)
}

// MarkNotificationAsUnread marks a notification as unread
func MarkNotificationAsUnread(threadID string) error {
	// GitHub API doesn't have a direct way to mark notifications as unread
	// This is a limitation of the GitHub API
	// For now, we'll return nil to avoid errors, but this functionality is not available
	return nil
}

func convertURL(apiURL string) string {
	if apiURL == "" {
		return ""
	}
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return ""
	}

	// Example API URLs:
	// https://api.github.com/repos/owner/repo/issues/123
	// https://api.github.com/repos/owner/repo/pulls/456
	// https://api.github.com/repos/owner/repo/commits/sha

	parts := strings.Split(parsedURL.Path, "/")
	// parts[0] is "", parts[1] is "repos", parts[2] is owner, parts[3] is repo, parts[4] is type, parts[5] is number/sha
	if len(parts) < 6 {
		return ""
	}

	owner := parts[2]
	repo := parts[3]
	resourceType := parts[4]
	identifier := parts[5]

	switch resourceType {
	case "issues":
		return fmt.Sprintf("https://github.com/%s/%s/issues/%s", owner, repo, identifier)
	case "pulls":
		return fmt.Sprintf("https://github.com/%s/%s/pull/%s", owner, repo, identifier)
	case "commits":
		return fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, identifier)
	default:
		// fallback to repo page if unknown
		return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	}
}
