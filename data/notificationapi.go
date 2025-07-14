package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
)

func GetNotifications(limit int, query ...string) ([]Notification, error) {
	return GetNotificationsPaginated(1, limit, query...)
}

func GetNotificationsPaginated(page, perPage int, query ...string) ([]Notification, error) {
	log.Debug("GetNotificationsPaginated", "page", page, "perPage", perPage, "query", query)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("GetNotificationsPaginated: failed to create client", "err", err)
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

	// Build query parameters
	params := url.Values{}
	params.Add("all", "true")
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("per_page", fmt.Sprintf("%d", perPage))

	endpoint := "notifications?" + params.Encode()
	log.Debug("GetNotificationsPaginated: calling API endpoint", "endpoint", endpoint)

	if err := client.Get(endpoint, &response); err != nil {
		log.Debug("GetNotificationsPaginated: API call failed", "err", err)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	log.Debug("GetNotificationsPaginated: received notifications", "count", len(response))

	// Handle potential empty response which could cause JSON parsing issues
	if response == nil {
		response = []struct {
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
		}{}
	}

	notifications := make([]Notification, 0, len(response))
	for _, n := range response {
		notification := Notification{
			ID:         n.ID,
			Title:      n.Subject.Title,
			Type:       n.Subject.Type,
			Repository: n.Repository.FullName,
			Reason:     Reason(n.Reason),
			Unread:     n.Unread,
			UpdatedAt:  n.UpdatedAt,
			URL:        convertURL(n.Subject.URL),
			ThreadID:   n.ID,
			Subscribed: true, // Default to subscribed (since we received the notification)
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func MarkNotificationAsRead(threadID string) error {
	log.Debug("MarkNotificationAsRead", "threadID", threadID)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("MarkNotificationAsRead: failed to create client", "err", err)
		return err
	}

	endpoint := fmt.Sprintf("notifications/threads/%s", threadID)
	log.Debug("MarkNotificationAsRead: calling PATCH", "endpoint", endpoint)

	// GitHub returns 205 Reset Content for successful mark-as-read, with no body
	// The Patch method expects a response to unmarshal, but GitHub returns empty body
	var response any
	err = client.Patch(endpoint, nil, &response)
	if err != nil {
		log.Error("MarkNotificationAsRead: PATCH failed", "err", err)
		return err
	}

	log.Debug("MarkNotificationAsRead: successfully marked as read", "response", response)

	return nil
}

func SubscribeForNotification(threadID string) error {
	log.Debug("SubscribeForNotification", "threadID", threadID)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("SubscribeForNotification: failed to create client", "err", err)
		return err
	}

	endpoint := fmt.Sprintf("notifications/threads/%s/subscription", threadID)
	log.Debug("SubscribeForNotification: calling PUT", "endpoint", endpoint)

	// TODO: add tags
	var response struct {
		Subscribed bool `json:"subscribed"`
		Ignored    bool
		Reason     string    `json:"reason"`
		CreatedAt  time.Time `json:"created_at"`
		URL        string
		ThreadURL  string
	}

	// GitHub returns 204 No Content for successful unsubscribe
	// TODO: refactor this?
	body := strings.NewReader("{\"ignored\":false}")
	err = client.Put(endpoint, body, &response)
	if err != nil {
		log.Error("SubscribeForNotification: PUT failed", "err", err)
		return err
	}

	if !response.Subscribed {
		log.Error("SubscribeForNotification, didn't managed to subscribed", "repsonse", response)
		return fmt.Errorf("subscribe didn't work")
	}

	log.Debug("SubscribeForNotification: successfully subscribed", "response", response)

	return nil
}

func UnsubscribeFromNotification(threadID string) error {
	log.Debug("UnsubscribeFromNotification", "threadID", threadID)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("UnsubscribeFromNotification: failed to create client", "err", err)
		return err
	}

	endpoint := fmt.Sprintf("notifications/threads/%s/subscription", threadID)
	log.Debug("UnsubscribeFromNotification: calling DELETE", "endpoint", endpoint)

	// GitHub returns 204 No Content for successful unsubscribe, with no body
	// Use nil response to avoid JSON parsing empty body
	err = client.Delete(endpoint, nil)
	if err != nil {
		log.Error("UnsubscribeFromNotification: DELETE failed", "err", err)
		return err
	}

	log.Debug("UnsubscribeFromNotification: successfully unsubscribed")

	return nil
}

// MarkNotificationAsDone marks a notification as done using the official API
func MarkNotificationAsDone(threadID string) error {
	log.Debug("MarkNotificationAsDone", "threadID", threadID)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("MarkNotificationAsDone: failed to create client", "err", err)
		return err
	}

	endpoint := fmt.Sprintf("notifications/threads/%s", threadID)
	log.Debug("MarkNotificationAsDone: calling DELETE", "endpoint", endpoint)

	// GitHub returns 204 No Content for successful mark-as-done, with no body
	// Use nil response to avoid JSON parsing empty body
	err = client.Delete(endpoint, nil)
	if err != nil {
		log.Debug("MarkNotificationAsDone: DELETE failed", "err", err)
		return err
	}

	log.Debug("MarkNotificationAsDone: successfully marked as done")
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
	// https://api.github.com/repos/owner/repo/discussions/123

	parts := strings.Split(parsedURL.Path, "/")
	// parts[0] is "", parts[1] is "repos", parts[2] is owner, parts[3] is repo, parts[4] is type, parts[5] is number/sha
	if len(parts) < 6 {
		// Handle shorter paths, fallback to repo page if we have at least owner/repo
		if len(parts) >= 4 {
			return fmt.Sprintf("https://github.com/%s/%s", parts[2], parts[3])
		}
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
		// Note: GitHub URLs use "pull" (singular) for pull requests
		return fmt.Sprintf("https://github.com/%s/%s/pull/%s", owner, repo, identifier)
	case "commits":
		return fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, identifier)
	case "discussions":
		return fmt.Sprintf("https://github.com/%s/%s/discussions/%s", owner, repo, identifier)
	case "releases":
		return fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", owner, repo, identifier)
	default:
		// fallback to repo page if unknown
		return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	}
}
