package utils

// FormatNotificationReason converts GitHub API notification reason codes to human-readable strings
func FormatNotificationReason(reason string) string {
	switch reason {
	case "review_requested":
		return "Review requested"
	case "mention":
		return "Mentioned"
	case "assign":
		return "Assigned"
	case "author":
		return "Author update"
	case "comment":
		return "New comment"
	case "ci_activity":
		return "CI activity"
	case "push":
		return "New push"
	case "team_mention":
		return "Team mentioned"
	case "state_change":
		return "State changed"
	case "subscribed":
		return "Subscribed"
	case "manual":
		return "Manual"
	case "security_alert":
		return "Security alert"
	case "your_activity":
		return "Your activity"
	default:
		// Return the original reason if we don't have a mapping
		// This ensures we don't lose information for new/unknown reasons
		return reason
	}
}
