package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

func TestNormalizeFilters(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"simple is:repo filter": {
			input:    "is:repo(owner/name)",
			expected: "repo:owner/name",
		},
		"is:repo with other filters": {
			input:    "reason:author is:repo(myorg/myrepo) unread:true",
			expected: "reason:author repo:myorg/myrepo unread:true",
		},
		"multiple is:repo filters": {
			input:    "is:repo(org1/repo1) is:repo(org2/repo2)",
			expected: "repo:org1/repo1 repo:org2/repo2",
		},
		"no is:repo filter": {
			input:    "reason:author unread:true",
			expected: "reason:author unread:true",
		},
		"existing repo filter unchanged": {
			input:    "repo:owner/name reason:author",
			expected: "repo:owner/name reason:author",
		},
		"mixed repo and is:repo filters": {
			input:    "repo:existing/repo is:repo(new/repo)",
			expected: "repo:existing/repo repo:new/repo",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := utils.NormalizeFilters(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestHasExplicitRepoFilter(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected bool
	}{
		"has repo: filter": {
			input:    "repo:owner/name reason:author",
			expected: true,
		},
		"has is:repo filter": {
			input:    "reason:author is:repo(owner/name)",
			expected: true,
		},
		"has both repo: and is:repo": {
			input:    "repo:existing/repo is:repo(new/repo)",
			expected: true,
		},
		"no repo filter": {
			input:    "reason:author unread:true",
			expected: false,
		},
		"repo word but not filter": {
			input:    "repository search terms",
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := utils.HasExplicitRepoFilter(tc.input)
			require.Equal(t, tc.expected, result, "Input: %s", tc.input)
		})
	}
}
