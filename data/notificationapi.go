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
	result, err := GetNotificationsPaginated(1, limit, query...)
	return result, err
}

func GetNotificationsWithLimits(limit int, maxLimit int, maxAgeDays int, query ...string) ([]Notification, error) {
	result, err := GetNotificationsPaginatedWithLimits(1, limit, maxLimit, maxAgeDays, query...)
	return result, err
}

func GetNotificationsPaginated(page, perPage int, query ...string) ([]Notification, error) {
	// Parse query to determine API parameters and client-side filters
	queryStr := ""
	if len(query) > 0 {
		queryStr = query[0]
	}

	// If there's a repo filter, we need to potentially fetch multiple pages
	// to get enough matching results
	if queryStr != "" && containsRepoFilter(queryStr) {
		result, err := getNotificationsWithRepoFilter(page, perPage, queryStr)
		return result, err
	}

	// For queries without repo filters, use the simpler single-page approach
	result, err := getNotificationsSinglePage(page, perPage, queryStr)
	return result, err
}

func GetNotificationsPaginatedWithCurrentCount(page, perPage int, maxLimit int, maxAgeDays int, currentCount int, query ...string) ([]Notification, error) {
	log.Debug("GetNotificationsPaginatedWithCurrentCount start", "page", page, "perPage", perPage, "maxLimit", maxLimit, "maxAgeDays", maxAgeDays, "currentCount", currentCount, "query", query)

	// Apply max limit enforcement based on actual current count (after deduplication)
	effectiveLimit := perPage
	if maxLimit > 0 {
		remainingLimit := maxLimit - currentCount
		if remainingLimit <= 0 {
			return []Notification{}, nil
		}
		if remainingLimit < perPage {
			effectiveLimit = remainingLimit
		}
	}

	// Parse query to determine API parameters and client-side filters
	queryStr := ""
	if len(query) > 0 {
		queryStr = query[0]
	}

	var notifications []Notification
	var err error

	// If there's a repo filter, we need to potentially fetch multiple pages
	// to get enough matching results
	if queryStr != "" && containsRepoFilter(queryStr) {
		notifications, err = getNotificationsWithRepoFilter(page, effectiveLimit, queryStr)
	} else {
		// For queries without repo filters, use the simpler single-page approach
		notifications, err = getNotificationsSinglePage(page, effectiveLimit, queryStr)
	}

	if err != nil {
		return nil, err
	}

	// Apply age filtering if specified
	if maxAgeDays > 0 {
		notifications = filterNotificationsByAge(notifications, maxAgeDays)
	}

	return notifications, nil
}

func GetNotificationsPaginatedWithLimits(page, perPage int, maxLimit int, maxAgeDays int, query ...string) ([]Notification, error) {
	log.Debug("GetNotificationsPaginatedWithLimits start", "page", page, "perPage", perPage, "maxLimit", maxLimit, "maxAgeDays", maxAgeDays, "query", query)

	// Apply max limit enforcement - ensure we never fetch more than maxLimit total
	effectiveLimit := perPage
	if maxLimit > 0 {
		totalRequested := (page-1)*perPage + perPage
		if totalRequested > maxLimit {
			remainingLimit := maxLimit - (page-1)*perPage
			if remainingLimit <= 0 {
				return []Notification{}, nil
			}
			effectiveLimit = remainingLimit
		}
	}

	// Parse query to determine API parameters and client-side filters
	queryStr := ""
	if len(query) > 0 {
		queryStr = query[0]
	}

	var notifications []Notification
	var err error

	// If there's a repo filter, we need to potentially fetch multiple pages
	// to get enough matching results
	if queryStr != "" && containsRepoFilter(queryStr) {
		notifications, err = getNotificationsWithRepoFilter(page, effectiveLimit, queryStr)
	} else {
		// For queries without repo filters, use the simpler single-page approach
		notifications, err = getNotificationsSinglePage(page, effectiveLimit, queryStr)
	}

	if err != nil {
		return nil, err
	}

	// Apply age filtering if specified
	if maxAgeDays > 0 {
		notifications = filterNotificationsByAge(notifications, maxAgeDays)
	}

	return notifications, nil
}

// getNotificationsSinglePage fetches a single page of notifications (original logic)
func getNotificationsSinglePage(page, perPage int, queryStr string) ([]Notification, error) {
	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("getNotificationsSinglePage: failed to create client", "err", err)
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

	// Handle is:unread filter by setting the 'all' parameter
	if strings.Contains(queryStr, "is:unread") {
		params.Add("all", "false")
	} else {
		params.Add("all", "true")
	}

	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("per_page", fmt.Sprintf("%d", perPage))

	endpoint := "notifications?" + params.Encode()
	log.Debug("getNotificationsSinglePage: calling API endpoint", "endpoint", endpoint)

	if err := client.Get(endpoint, &response); err != nil {
		log.Debug("getNotificationsSinglePage: API call failed", "err", err)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	log.Debug("getNotificationsSinglePage: received notifications", "count", len(response))

	// Handle potential empty response
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
			Subscribed: true,
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// getNotificationsWithRepoFilter fetches multiple pages if needed to fill the requested limit
func getNotificationsWithRepoFilter(requestedPage, perPage int, queryStr string) ([]Notification, error) {
	log.Debug("getNotificationsWithRepoFilter: starting multi-page fetch", "requestedPage", requestedPage, "perPage", perPage)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("getNotificationsWithRepoFilter: failed to create client", "err", err)
		return nil, err
	}

	var filteredNotifications []Notification
	currentPage := 1
	maxPages := 10 // Safety limit to prevent infinite loops

	if requestedPage == 1 {
		for len(filteredNotifications) < perPage && currentPage <= maxPages {
			// Fetch this page
			notifications, hasMore, err := fetchSingleNotificationPage(client, currentPage, perPage, queryStr)
			if err != nil {
				return nil, err
			}

			// Filter the notifications from this page
			filteredFromThisPage := filterNotificationsByRepo(notifications, queryStr)
			filteredNotifications = append(filteredNotifications, filteredFromThisPage...)

			// If this page didn't return a full set of results, we've reached the end
			if !hasMore {
				break
			}

			currentPage++
		}

		// Return up to perPage results
		if len(filteredNotifications) > perPage {
			result := filteredNotifications[:perPage]
			return result, nil
		}
		return filteredNotifications, nil
	}

	// For pages > 1, fall back to single page for now
	notifications, _, err := fetchSingleNotificationPage(client, requestedPage, perPage, queryStr)
	if err != nil {
		return nil, err
	}

	result := filterNotificationsByRepo(notifications, queryStr)
	return result, nil
}

// fetchSingleNotificationPage fetches a single page and returns whether there are more pages
func fetchSingleNotificationPage(client *gh.RESTClient, page, perPage int, queryStr string) ([]Notification, bool, error) {
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

	// Handle is:unread filter by setting the 'all' parameter
	if strings.Contains(queryStr, "is:unread") {
		params.Add("all", "false")
	} else {
		params.Add("all", "true")
	}

	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("per_page", fmt.Sprintf("%d", perPage))

	endpoint := "notifications?" + params.Encode()

	if err := client.Get(endpoint, &response); err != nil {
		return nil, false, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Handle potential empty response
	if response == nil {
		return []Notification{}, false, nil
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
			Subscribed: true,
		}

		notifications = append(notifications, notification)
	}

	// If we got exactly perPage results, assume there might be more pages
	hasMore := len(response) == perPage

	return notifications, hasMore, nil
}

// containsRepoFilter checks if the query contains a repo filter
func containsRepoFilter(query string) bool {
	tokens := strings.Fields(query)
	for _, token := range tokens {
		if strings.HasPrefix(token, "repo:") {
			return true
		}
	}
	return false
}

// filterNotificationsByRepo applies repo filtering to notifications
func filterNotificationsByRepo(notifications []Notification, query string) []Notification {
	// Extract repo filter from query
	repoFilter := ""
	tokens := strings.Fields(query)
	for _, token := range tokens {
		if strings.HasPrefix(token, "repo:") {
			repoFilter = strings.TrimPrefix(token, "repo:")
			break
		}
	}

	if repoFilter == "" {
		return notifications
	}

	// Filter notifications by repository
	filtered := make([]Notification, 0, len(notifications))
	for _, notification := range notifications {
		if strings.Contains(notification.Repository, repoFilter) {
			filtered = append(filtered, notification)
		}
	}

	return filtered
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
	if err != nil && err.Error() != "unexpected end of JSON input" {
		log.Error("MarkNotificationAsRead: PATCH failed", "err", err)
		return err
	}

	log.Debug("MarkNotificationAsRead: successfully marked as read")

	return nil
}

// Removed unused, commented-out functions for subscribe/unsubscribe

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

// filterNotificationsByAge filters notifications based on maximum age in days
func filterNotificationsByAge(notifications []Notification, maxAgeDays int) []Notification {
	if maxAgeDays <= 0 {
		return notifications
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)
	filtered := make([]Notification, 0, len(notifications))

	for _, notification := range notifications {
		if notification.UpdatedAt.After(cutoffTime) {
			filtered = append(filtered, notification)
		}
	}

	log.Debug("filterNotificationsByAge", "original", len(notifications), "filtered", len(filtered), "maxAgeDays", maxAgeDays)
	return filtered
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
