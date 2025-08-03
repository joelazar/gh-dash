package data

import (
	"testing"
	"time"
)

func TestDeduplicateNotifications(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	latest := now.Add(1 * time.Hour)

	tests := map[string]struct {
		input    []Notification
		expected []Notification
	}{
		"no duplicates": {
			input: []Notification{
				{
					ID:         "1",
					Title:      "Different Title 1",
					Type:       "Issue",
					Repository: "owner/repo1",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Different Title 2",
					Type:       "PullRequest",
					Repository: "owner/repo2",
					Reason:     "review_requested",
					UpdatedAt:  now,
				},
			},
			expected: []Notification{
				{
					ID:         "1",
					Title:      "Different Title 1",
					Type:       "Issue",
					Repository: "owner/repo1",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Different Title 2",
					Type:       "PullRequest",
					Repository: "owner/repo2",
					Reason:     "review_requested",
					UpdatedAt:  now,
				},
			},
		},
		"exact duplicates - keep latest": {
			input: []Notification{
				{
					ID:         "1",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  earlier,
				},
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  latest,
				},
				{
					ID:         "3",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  now,
				},
			},
			expected: []Notification{
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  latest,
				},
			},
		},
		"different repos, same other fields": {
			input: []Notification{
				{
					ID:         "1",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo1",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo2",
					Reason:     "mention",
					UpdatedAt:  now,
				},
			},
			expected: []Notification{
				{
					ID:         "1",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo1",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo2",
					Reason:     "mention",
					UpdatedAt:  now,
				},
			},
		},
		"different reasons, same other fields": {
			input: []Notification{
				{
					ID:         "1",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "assign",
					UpdatedAt:  now,
				},
			},
			expected: []Notification{
				{
					ID:         "1",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "mention",
					UpdatedAt:  now,
				},
				{
					ID:         "2",
					Title:      "Same Title",
					Type:       "Issue",
					Repository: "owner/repo",
					Reason:     "assign",
					UpdatedAt:  now,
				},
			},
		},
		"empty slice": {
			input:    []Notification{},
			expected: []Notification{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := DeduplicateNotifications(test.input)
			
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d notifications, got %d", len(test.expected), len(result))
				return
			}

			// Create maps for easier comparison
			resultMap := make(map[string]Notification)
			expectedMap := make(map[string]Notification)
			
			for _, n := range result {
				resultMap[n.ID] = n
			}
			for _, n := range test.expected {
				expectedMap[n.ID] = n
			}

			for id, expected := range expectedMap {
				actual, exists := resultMap[id]
				if !exists {
					t.Errorf("Expected notification with ID %s not found in result", id)
					continue
				}
				
				if actual.Title != expected.Title ||
					actual.Type != expected.Type ||
					actual.Repository != expected.Repository ||
					actual.Reason != expected.Reason ||
					!actual.UpdatedAt.Equal(expected.UpdatedAt) {
					t.Errorf("Notification %s differs from expected", id)
				}
			}
		})
	}
}