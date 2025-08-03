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

func GetNotificationsPaginatedWithLimits(page, perPage int, maxLimit int, maxAgeDays int, query ...string) ([]Notification, error) {
	log.Debug("GetNotificationsPaginatedWithLimits start", "page", page, "perPage", perPage, "maxLimit", maxLimit, "maxAgeDays", maxAgeDays, "query", query)
	start := time.Now()

	// Apply max limit enforcement - ensure we never fetch more than maxLimit total
	effectiveLimit := perPage
	if maxLimit > 0 {
		totalRequested := (page-1)*perPage + perPage
		log.Debug("PERF: Limit calculation - totalRequested: %d, maxLimit: %d", totalRequested, maxLimit)
		if totalRequested > maxLimit {
			remainingLimit := maxLimit - (page-1)*perPage
			log.Debug("PERF: Remaining limit: %d", remainingLimit)
			if remainingLimit <= 0 {
				log.Debug("PERF: No remaining limit, returning empty")
				return []Notification{}, nil
			}
			effectiveLimit = remainingLimit
			log.Debug("PERF: Effective limit adjusted to: %d", effectiveLimit)
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
		log.Debug("PERF: Using repo filter path with effectiveLimit: %d", effectiveLimit)
		notifications, err = getNotificationsWithRepoFilter(page, effectiveLimit, queryStr)
	} else {
		// For queries without repo filters, use the simpler single-page approach
		log.Debug("PERF: Using single page path with effectiveLimit: %d", effectiveLimit)
		notifications, err = getNotificationsSinglePage(page, effectiveLimit, queryStr)
	}

	if err != nil {
		return nil, err
	}

	// Apply age filtering if specified
	if maxAgeDays > 0 {
		preFilterCount := len(notifications)
		notifications = filterNotificationsByAge(notifications, maxAgeDays)
		log.Debug("PERF: Age filtering - before: %d, after: %d, maxAgeDays: %d", preFilterCount, len(notifications), maxAgeDays)
	}

	log.Debug("PERF: GetNotificationsPaginatedWithLimits COMPLETE in %v - returned %d notifications", time.Since(start), len(notifications))
	return notifications, nil
}

// getNotificationsSinglePage fetches a single page of notifications (original logic)
func getNotificationsSinglePage(page, perPage int, queryStr string) ([]Notification, error) {
	log.Debug("PERF: getNotificationsSinglePage START - page: %d, perPage: %d, query: '%s'", page, perPage, queryStr)
	start := time.Now()

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("PERF: getNotificationsSinglePage: failed to create client in %v - error: %v", time.Since(start), err)
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
	log.Debug("PERF: getNotificationsSinglePage: calling API endpoint: %s", endpoint)
	log.Debug("getNotificationsSinglePage: calling API endpoint", "endpoint", endpoint)

	apiStart := time.Now()
	if err := client.Get(endpoint, &response); err != nil {
		log.Debug("PERF: getNotificationsSinglePage: API call FAILED in %v - error: %v", time.Since(apiStart), err)
		log.Debug("getNotificationsSinglePage: API call failed", "err", err)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	log.Debug("PERF: getNotificationsSinglePage: API call SUCCESS in %v", time.Since(apiStart))

	log.Debug("PERF: getNotificationsSinglePage: received %d raw notifications from API", len(response))
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

	processStart := time.Now()
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
	log.Debug("PERF: getNotificationsSinglePage: processed %d notifications in %v", len(notifications), time.Since(processStart))

	log.Debug("PERF: getNotificationsSinglePage COMPLETE in %v - processed %d notifications", time.Since(start), len(notifications))
	return notifications, nil
}

// getNotificationsWithRepoFilter fetches multiple pages if needed to fill the requested limit
func getNotificationsWithRepoFilter(requestedPage, perPage int, queryStr string) ([]Notification, error) {
	log.Debug("PERF: getNotificationsWithRepoFilter START - requestedPage: %d, perPage: %d, query: '%s'", requestedPage, perPage, queryStr)
	start := time.Now()
	log.Debug("getNotificationsWithRepoFilter: starting multi-page fetch", "requestedPage", requestedPage, "perPage", perPage)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("getNotificationsWithRepoFilter: failed to create client", "err", err)
		return nil, err
	}

	// For simplicity and performance, we'll optimize for the common case (page 1)
	// and fetch multiple pages until we have enough results
	var filteredNotifications []Notification
	currentPage := 1
	maxPages := 10 // Safety limit to prevent infinite loops
	log.Debug("PERF: getNotificationsWithRepoFilter: maxPages limit set to %d", maxPages)

	// For page 1, fetch until we have enough results or run out of pages
	if requestedPage == 1 {
		log.Debug("PERF: getNotificationsWithRepoFilter: handling page 1 - need %d filtered results", perPage)
		for len(filteredNotifications) < perPage && currentPage <= maxPages {
			log.Debug("PERF: getNotificationsWithRepoFilter: fetching page %d (have %d/%d filtered so far)", currentPage, len(filteredNotifications), perPage)
			// Fetch this page
			pageStart := time.Now()
			notifications, hasMore, err := fetchSingleNotificationPage(client, currentPage, perPage, queryStr)
			if err != nil {
				log.Debug("PERF: getNotificationsWithRepoFilter: page %d fetch FAILED in %v - error: %v", currentPage, time.Since(pageStart), err)
				return nil, err
			}
			log.Debug("PERF: getNotificationsWithRepoFilter: page %d fetch SUCCESS in %v - got %d raw notifications", currentPage, time.Since(pageStart), len(notifications))

			// Filter the notifications from this page
			filterStart := time.Now()
			filteredFromThisPage := filterNotificationsByRepo(notifications, queryStr)
			log.Debug("PERF: getNotificationsWithRepoFilter: page %d filtering took %v - %d->%d notifications", currentPage, time.Since(filterStart), len(notifications), len(filteredFromThisPage))
			filteredNotifications = append(filteredNotifications, filteredFromThisPage...)

			log.Debug("PERF: getNotificationsWithRepoFilter: page %d results - raw: %d, filtered: %d, total_filtered: %d, hasMore: %v", currentPage, len(notifications), len(filteredFromThisPage), len(filteredNotifications), hasMore)
			log.Debug("getNotificationsWithRepoFilter: page results",
				"page", currentPage,
				"raw", len(notifications),
				"filtered", len(filteredFromThisPage),
				"total_filtered", len(filteredNotifications))

			// If this page didn't return a full set of results, we've reached the end
			if !hasMore {
				log.Debug("PERF: getNotificationsWithRepoFilter: no more pages available, stopping at page %d", currentPage)
				break
			}

			currentPage++
		}

		log.Debug("PERF: getNotificationsWithRepoFilter: page 1 processing complete - got %d filtered notifications", len(filteredNotifications))
		// Return up to perPage results
		if len(filteredNotifications) > perPage {
			log.Debug("PERF: getNotificationsWithRepoFilter: trimming %d results to %d", len(filteredNotifications), perPage)
			result := filteredNotifications[:perPage]
			log.Debug("PERF: getNotificationsWithRepoFilter COMPLETE (page 1, trimmed) in %v - returning %d notifications", time.Since(start), len(result))
			return result, nil
		}
		log.Debug("PERF: getNotificationsWithRepoFilter COMPLETE (page 1, full) in %v - returning %d notifications", time.Since(start), len(filteredNotifications))
		return filteredNotifications, nil
	}

	// For pages > 1, fall back to single page for now (this is complex to implement efficiently)
	log.Debug("PERF: getNotificationsWithRepoFilter: handling page %d (>1) - using fallback single page approach", requestedPage)
	pageStart := time.Now()
	notifications, _, err := fetchSingleNotificationPage(client, requestedPage, perPage, queryStr)
	if err != nil {
		log.Debug("PERF: getNotificationsWithRepoFilter: page %d fetch FAILED in %v - error: %v", requestedPage, time.Since(pageStart), err)
		return nil, err
	}
	log.Debug("PERF: getNotificationsWithRepoFilter: page %d fetch SUCCESS in %v - got %d raw notifications", requestedPage, time.Since(pageStart), len(notifications))

	filterStart := time.Now()
	result := filterNotificationsByRepo(notifications, queryStr)
	log.Debug("PERF: getNotificationsWithRepoFilter: page %d filtering took %v - %d->%d notifications", requestedPage, time.Since(filterStart), len(notifications), len(result))
	log.Debug("PERF: getNotificationsWithRepoFilter COMPLETE (page %d) in %v - returning %d notifications", requestedPage, time.Since(start), len(result))
	return result, nil
}

// fetchSingleNotificationPage fetches a single page and returns whether there are more pages
func fetchSingleNotificationPage(client *gh.RESTClient, page, perPage int, queryStr string) ([]Notification, bool, error) {
	log.Debug("PERF: fetchSingleNotificationPage START - page: %d, perPage: %d, query: '%s'", page, perPage, queryStr)
	start := time.Now()
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
	log.Debug("PERF: fetchSingleNotificationPage: calling API endpoint: %s", endpoint)

	apiStart := time.Now()
	if err := client.Get(endpoint, &response); err != nil {
		log.Debug("PERF: fetchSingleNotificationPage: API call FAILED in %v - error: %v", time.Since(apiStart), err)
		return nil, false, fmt.Errorf("failed to get notifications: %w", err)
	}
	log.Debug("PERF: fetchSingleNotificationPage: API call SUCCESS in %v", time.Since(apiStart))

	// Handle potential empty response
	if response == nil {
		log.Debug("PERF: fetchSingleNotificationPage: received nil response")
		return []Notification{}, false, nil
	}
	log.Debug("PERF: fetchSingleNotificationPage: received %d raw notifications from API", len(response))

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
	log.Debug("PERF: fetchSingleNotificationPage: processed %d notifications, hasMore: %v", len(notifications), hasMore)

	log.Debug("PERF: fetchSingleNotificationPage COMPLETE in %v - returning %d notifications, hasMore: %v", time.Since(start), len(notifications), hasMore)
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
	log.Debug("PERF: filterNotificationsByRepo START - filtering %d notifications with query: '%s'", len(notifications), query)
	start := time.Now()

	// Extract repo filter from query
	repoFilter := ""
	tokens := strings.Fields(query)
	for _, token := range tokens {
		if strings.HasPrefix(token, "repo:") {
			repoFilter = strings.TrimPrefix(token, "repo:")
			break
		}
	}
	log.Debug("PERF: filterNotificationsByRepo: extracted repo filter: '%s'", repoFilter)

	if repoFilter == "" {
		log.Debug("PERF: filterNotificationsByRepo: no repo filter, returning all %d notifications", len(notifications))
		return notifications
	}

	// Filter notifications by repository
	filtered := make([]Notification, 0, len(notifications))
	matchCount := 0
	for _, notification := range notifications {
		if strings.Contains(notification.Repository, repoFilter) {
			filtered = append(filtered, notification)
			matchCount++
		}
	}
	log.Debug("PERF: filterNotificationsByRepo: matched %d/%d notifications", matchCount, len(notifications))

	log.Debug("PERF: filterNotificationsByRepo COMPLETE in %v - filtered %d->%d notifications", time.Since(start), len(notifications), len(filtered))
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
	// TODO: this one is not working, fix it or remove it
	// if err != nil && !errors.Is(err, json.SyntaxError) {
	if err != nil && err.Error() != "unexpected end of JSON input" {
		log.Error("MarkNotificationAsRead: PATCH failed", "err", err)
		return err
	}

	log.Debug("MarkNotificationAsRead: successfully marked as read")

	return nil
}

/* func SubscribeForNotification(threadID string) error {
	log.Debug("SubscribeForNotification", "threadID", threadID)

	client, err := gh.DefaultRESTClient()
	if err != nil {
		log.Debug("SubscribeForNotification: failed to create client", "err", err)
		return err
	}

	endpoint := fmt.Sprintf("notifications/threads/%s/subscription", threadID)
	log.Debug("SubscribeForNotification: calling PUT", "endpoint", endpoint)

	var response struct {
		Subscribed bool      `json:"subscribed"`
		Ignored    bool      `json:"ignored"`
		Reason     string    `json:"reason"`
		CreatedAt  time.Time `json:"created_at"`
		URL        string    `json:"url"`
		ThreadURL  string    `json:"thread_url"`
	}

	// GitHub returns 200 OK with a response body for successful subscribe.
	body := strings.NewReader("{\"ignored\": false}")
	err = client.Put(endpoint, body, &response)
	if err != nil {
		log.Error("SubscribeForNotification: PUT failed", "err", err)
		return err
	}

	if !response.Subscribed {
		log.Error("SubscribeForNotification: failed to subscribe", "response", response)
		return fmt.Errorf("failed to subscribe to notification thread")
	}

	log.Debug("SubscribeForNotification: successfully subscribed", "response", response)

	return nil
} */

/* func UnsubscribeFromNotification(threadID string) error {
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
} */

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
