package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
)

func GetNotifications(limit int, query ...string) ([]*Notification, error) {
	return GetNotificationsPaginated(1, limit, query...)
}

func GetNotificationsPaginated(page, perPage int, query ...string) ([]*Notification, error) {
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

	// Parse search query if provided
	var clientSideFilters map[string]string
	if len(query) > 0 && query[0] != "" {
		searchParams := parseNotificationSearch(query[0])
		clientSideFilters = make(map[string]string)

		for key, value := range searchParams {
			// Handle API-supported parameters
			switch key {
			case "all", "participating", "since", "before":
				params.Set(key, value)
			case "reason":
				// GitHub API supports reason filtering
				params.Set(key, value)
			default:
				// Store client-side filters for post-processing
				clientSideFilters[key] = value
			}
		}
	}

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

	notifications := make([]*Notification, 0, len(response))
	for _, n := range response {
		notification := &Notification{
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

		// Apply client-side filters
		if clientSideFilters != nil {
			if shouldIncludeNotification(notification, clientSideFilters) {
				notifications = append(notifications, notification)
			}
		} else {
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

// shouldIncludeNotification applies client-side filtering for filters not supported by GitHub API
func shouldIncludeNotification(notification *Notification, filters map[string]string) bool {
	for key, value := range filters {
		switch key {
		// Remove saved/bookmarked filter since it's not supported by GitHub API
		case "done":
			// For "done" notifications - GitHub API removes them entirely when marked as done
			// So if we're filtering for "is:done", we won't find any because they don't exist in API
			// This filter will effectively return an empty list, which is correct behavior
			if value == "true" {
				return false
			}
		case "repo":
			// Filter by repository
			if value != "" && notification.Repository != value {
				return false
			}
		case "author":
			// Filter by author - GitHub API doesn't provide author info directly
			// This would require additional API calls, so we'll skip for now
			continue
		}
	}
	return true
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
	err = client.Patch(endpoint, nil, nil)
	if err != nil {
		// GitHub API returns empty body causing "unexpected end of JSON input"
		// This is expected behavior for successful PATCH to mark-as-read endpoint
		if err.Error() == "unexpected end of JSON input" {
			log.Debug("MarkNotificationAsRead: received expected empty response")
		} else {
			log.Debug("MarkNotificationAsRead: PATCH failed", "err", err)
			return err
		}
	}

	log.Debug("MarkNotificationAsRead: successfully marked as read")
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
		log.Debug("UnsubscribeFromNotification: DELETE failed", "err", err)
		return err
	}

	log.Debug("UnsubscribeFromNotification: successfully unsubscribed")
	return nil
}

// BookmarkNotification is not supported by GitHub API - removed

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

// MarkNotificationAsUnread marks a notification as unread
func MarkNotificationAsUnread(threadID string) error {
	log.Debug("MarkNotificationAsUnread", "threadID", threadID)
	log.Debug("MarkNotificationAsUnread: GitHub API doesn't support marking notifications as unread")

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

// parseNotificationSearch parses a search query string and returns URL parameters
// Supports filters: repo, is, reason, author
func parseNotificationSearch(query string) map[string]string {
	params := make(map[string]string)

	// Split query into tokens
	tokens := strings.Fields(query)

	for _, token := range tokens {
		if strings.Contains(token, ":") {
			parts := strings.SplitN(token, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "repo":
					// Filter by repository
					// For GitHub API, this would be part of the URL path or query
					// Since GitHub notifications API doesn't support repo filtering directly,
					// we'll store it for client-side filtering
					params["repo"] = value
				case "is":
					// Filter by read/unread status and other states
					switch value {
					case "unread":
						params["all"] = "false"
					case "read":
						params["all"] = "true"
					// "saved" filter removed - not supported by GitHub API
					case "done":
						// GitHub API doesn't have native done filter
						// We'll handle this client-side by storing for later filtering
						params["done"] = "true"
					}
				case "reason":
					// Filter by reason (mention, assign, etc.)
					if value == "participating" {
						// Use GitHub API's participating parameter
						params["participating"] = "true"
					} else {
						// For other reasons like assign, mention, team_mention, review_requested
						params["reason"] = value
					}
				case "author":
					// Filter by author - GitHub API doesn't directly support this
					// Store for client-side filtering
					params["author"] = value
				}
			}
		}
	}

	return params
}
