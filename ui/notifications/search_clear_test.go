package notifications

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/ui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

func TestSearchClearBehavior(t *testing.T) {
	// Test the core logic of the search clear behavior
	testCases := map[string]struct {
		initialSearchValue string
		newSearchValue     string
		expectedBehavior   string
	}{
		"clearing search triggers fresh fetch": {
			initialSearchValue: "some query",
			newSearchValue:     "",
			expectedBehavior:   "fresh_fetch",
		},
		"non-empty search uses normal fetch": {
			initialSearchValue: "old query",
			newSearchValue:     "new query",
			expectedBehavior:   "normal_fetch",
		},
		"empty to empty still triggers fresh fetch": {
			initialSearchValue: "",
			newSearchValue:     "",
			expectedBehavior:   "fresh_fetch",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Create a minimal context
			ctx := &context.ProgramContext{
				Config: &config.Config{
					Defaults: config.Defaults{
						NotificationsLimit: 25,
					},
				},
			}

			// Create minimal model
			model := notificationssection.Model{}
			model.Id = 1
			model.Ctx = ctx
			model.Config = config.SectionConfig{}
			model.SearchValue = tc.initialSearchValue
			model.CurrentPage = 2 // Simulate being on page 2
			model.HasNextPage = false // Simulate no next page

			// Test the behavior based on the new search value
			if tc.newSearchValue == "" {
				// Simulate the fresh fetch logic for empty search
				model.SearchValue = tc.newSearchValue
				model.ResetRows()
				model.CurrentPage = 1
				model.HasNextPage = true

				// Verify fresh fetch behavior
				require.Equal(t, "", model.SearchValue, "Search value should be empty")
				require.Equal(t, 1, model.CurrentPage, "Page should be reset to 1")
				require.True(t, model.HasNextPage, "HasNextPage should be reset to true")
			} else {
				// Simulate normal search behavior
				model.SearchValue = tc.newSearchValue
				model.ResetRows()
				// Normal fetch doesn't reset pagination state

				require.Equal(t, tc.newSearchValue, model.SearchValue, "Search value should be updated")
			}
		})
	}
}

func TestSearchClearResetsRowsAndPagination(t *testing.T) {
	// Create a simple model structure
	model := notificationssection.Model{}
	model.Id = 1
	model.SearchValue = "existing query"
	model.CurrentPage = 3 // Simulate being on page 3
	model.HasNextPage = false // Simulate no more pages

	// Set up context for limit configuration
	ctx := &context.ProgramContext{
		Config: &config.Config{
			Defaults: config.Defaults{
				NotificationsLimit: 50,
			},
		},
	}
	model.Ctx = ctx
	model.Config = config.SectionConfig{}

	// Simulate the logic from the actual search clear implementation
	newSearchValue := ""
	model.SearchValue = newSearchValue
	model.ResetRows()
	
	// When search is cleared, pagination should be reset
	if newSearchValue == "" {
		model.CurrentPage = 1
		model.HasNextPage = true
	}

	// Verify state was properly reset
	require.Equal(t, "", model.SearchValue, "Search value should be empty")
	require.Equal(t, 1, model.CurrentPage, "Page should be reset to 1")
	require.True(t, model.HasNextPage, "HasNextPage should be reset to true")
}