package utils

import (
	"regexp"
	"strings"
)

// NormalizeFilters converts is:repo(<name>) syntax to repo:<name> and ensures
// that explicit repo filters prevent smart-repo filtering from being applied
func NormalizeFilters(filters string) string {
	// Regular expression to match is:repo(<name>) pattern
	isRepoPattern := regexp.MustCompile(`is:repo\(([^)]+)\)`)
	
	// Replace is:repo(<name>) with repo:<name>
	normalized := isRepoPattern.ReplaceAllString(filters, "repo:$1")
	
	return normalized
}

// HasExplicitRepoFilter checks if the filters contain an explicit repo filter
// (either repo:<name> or is:repo(<name>))
func HasExplicitRepoFilter(filters string) bool {
	// Check for repo: prefix
	for _, token := range strings.Fields(filters) {
		if strings.HasPrefix(token, "repo:") {
			return true
		}
	}
	
	// Check for is:repo(<name>) pattern
	isRepoPattern := regexp.MustCompile(`is:repo\([^)]+\)`)
	return isRepoPattern.MatchString(filters)
}