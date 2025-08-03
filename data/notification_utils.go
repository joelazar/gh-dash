package data

import (
	"fmt"

	"github.com/charmbracelet/log"
)

// DeduplicateNotifications removes duplicate notifications based on reason, type, repository, and title
// keeping only the latest one (by UpdatedAt) for each unique combination
func DeduplicateNotifications(notifications []Notification) []Notification {
	if len(notifications) == 0 {
		return notifications
	}

	// Create a map to track the latest notification for each unique combination
	// Key: combination of reason, type, repository, and title
	// Value: index of the latest notification in the original slice
	latestMap := make(map[string]int)
	
	for i, notification := range notifications {
		// Create a unique key from the fields we want to deduplicate on
		key := fmt.Sprintf("%s|%s|%s|%s", 
			string(notification.Reason), 
			notification.Type, 
			notification.Repository, 
			notification.Title)
		
		// Check if we've seen this combination before
		if existingIndex, exists := latestMap[key]; exists {
			// Keep the one with the later UpdatedAt timestamp
			if notification.UpdatedAt.After(notifications[existingIndex].UpdatedAt) {
				latestMap[key] = i
			}
		} else {
			// First time seeing this combination
			latestMap[key] = i
		}
	}
	
	// Collect all the unique (latest) notifications
	uniqueNotifications := make([]Notification, 0, len(latestMap))
	addedIndices := make(map[int]bool)
	
	// Preserve original order by iterating through original slice
	for i, notification := range notifications {
		if addedIndices[i] {
			continue
		}
		
		key := fmt.Sprintf("%s|%s|%s|%s", 
			string(notification.Reason), 
			notification.Type, 
			notification.Repository, 
			notification.Title)
		
		if latestIndex, exists := latestMap[key]; exists && latestIndex == i {
			uniqueNotifications = append(uniqueNotifications, notification)
			addedIndices[i] = true
		}
	}
	
	log.Debug("deduplicateNotifications", "original", len(notifications), "deduplicated", len(uniqueNotifications))
	return uniqueNotifications
}