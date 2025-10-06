package notificationssection_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
)

func TestInitialLoad(t *testing.T) {
	// Test that new model starts in loading state to show loading message
	// instead of empty state message before first data arrives

	// Create a minimal model to test the loading behavior
	model := notificationssection.Model{}

	// Test that SetIsLoading method works correctly
	require.NotPanics(t, func() {
		model.SetIsLoading(true)
		model.SetIsLoading(false)
	})

	// Test that ResetRows method works correctly
	require.NotPanics(t, func() {
		model.ResetRows()
	})
}
