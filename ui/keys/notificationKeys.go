package keys

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	log "github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
)

type NotificationKeyMap struct {
	ToggleSubscription key.Binding
	MarkDone           key.Binding
	ToggleRead         key.Binding
	OpenBrowser        key.Binding
	ViewSwitch         key.Binding
}

var NotificationKeys = NotificationKeyMap{
	ToggleSubscription: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("S", "toggle subscription/unsubscription"),
	),
	MarkDone: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "mark as done"),
	),
	ToggleRead: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle read/unread"),
	),
	OpenBrowser: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
	ViewSwitch: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "switch view"),
	),
}

func NotificationFullHelp() []key.Binding {
	return []key.Binding{
		NotificationKeys.ToggleSubscription,
		NotificationKeys.MarkDone,
		NotificationKeys.ToggleRead,
		NotificationKeys.OpenBrowser,
		NotificationKeys.ViewSwitch,
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
		case "toggleSubscription":
			key = &NotificationKeys.ToggleSubscription
		case "markDone":
			key = &NotificationKeys.MarkDone
		case "toggleRead":
			key = &NotificationKeys.ToggleRead
		case "openBrowser":
			key = &NotificationKeys.OpenBrowser
		case "viewSwitch":
			key = &NotificationKeys.ViewSwitch
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
