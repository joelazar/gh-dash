package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
)

type NotificationKeyMap struct {
	MarkDone   key.Binding
	MarkRead   key.Binding
	ViewSwitch key.Binding
	SortToggle key.Binding
}

var NotificationKeys = NotificationKeyMap{
	MarkRead: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "mark as read"),
	),
	MarkDone: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "mark as done"),
	),
	ViewSwitch: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch view"),
	),
	SortToggle: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "toggle sort"),
	),
}

func NotificationFullHelp() []key.Binding {
	return []key.Binding{
		NotificationKeys.MarkRead,
		NotificationKeys.MarkDone,
		NotificationKeys.ViewSwitch,
		NotificationKeys.SortToggle,
	}
}

// CustomNotificationBindings stores custom keybindings that don't have built-in equivalents
var CustomNotificationBindings []key.Binding

func rebindNotificationKeys(keys []config.Keybinding) error {
	CustomNotificationBindings = []key.Binding{}

	for _, notificationKey := range keys {
		if notificationKey.Builtin == "" {
			// Handle custom commands
			if notificationKey.Command != "" {
				name := notificationKey.Name
				if notificationKey.Name == "" {
					name = config.TruncateCommand(notificationKey.Command)
				}

				customBinding := key.NewBinding(
					key.WithKeys(notificationKey.Key),
					key.WithHelp(notificationKey.Key, name),
				)

				CustomNotificationBindings = append(CustomNotificationBindings, customBinding)
			}
			continue
		}

		log.Debug("Rebinding notification key", "builtin", notificationKey.Builtin, "key", notificationKey.Key)

		var key *key.Binding

		switch notificationKey.Builtin {
		case "markRead":
			key = &NotificationKeys.MarkRead
		case "markDone":
			key = &NotificationKeys.MarkDone
		case "viewSwitch":
			key = &NotificationKeys.ViewSwitch
		case "sortToggle":
			key = &NotificationKeys.SortToggle
		default:
			return fmt.Errorf("unknown built-in notification key: '%s'", notificationKey.Builtin)
		}

		key.SetKeys(notificationKey.Key)

		helpDesc := key.Help().Desc
		if notificationKey.Name != "" {
			helpDesc = notificationKey.Name
		}
		key.SetHelp(notificationKey.Key, helpDesc)
	}

	return nil
}
