---
title: Notifications Keybindings
linkTitle: >-
  ![icon:bell](lucide)&nbsp;Notifications
summary: >-
  Define keybindings for the Notifications view.
weight: 3
---

You can define custom keybindings for the Notifications view using the `keybindings.notifications` setting in your [configuration file][01].

## Schema Reference

Each keybinding entry in the `notifications` array must have:

- `key`: The keystroke combination to trigger the action
- Either `builtin` for built-in actions or `command` for custom commands
- `name`: A description of what the action does (optional for built-in actions)

## Built-in Actions

| Action       | Description                                                |
| :----------- | :--------------------------------------------------------- |
| `markRead`   | Mark the selected notification as read                     |
| `markDone`   | Mark the selected notification as done (read and archived) |
| `viewSwitch` | Switch between different dashboard views                   |

## Template Variables

When defining custom commands, you can use these template variables from the selected notification:

| Variable                    | Description                                               |
| :-------------------------- | :-------------------------------------------------------- |
| `{{.Title}}`                | The notification title                                    |
| `{{.GetRepoNameWithOwner}}` | Repository name with owner (e.g., `owner/repo`)           |
| `{{.URL}}`                  | The notification URL                                      |
| `{{.ThreadID}}`             | The notification thread ID                                |
| `{{.Type}}`                 | Notification type (e.g., `PullRequest`, `Issue`)          |
| `{{.Reason}}`               | Notification reason (e.g., `review_requested`, `mention`) |
| `{{.Repository}}`           | Repository name                                           |
| `{{.Unread}}`               | Whether the notification is unread                        |

## Examples

### Using Built-in Actions

```yaml
keybindings:
  notifications:
    - key: "m"
      builtin: "markRead"
      name: "Mark as read"
    - key: "d"
      builtin: "markDone"
      name: "Mark as done"
    - key: "ctrl+r"
      builtin: "markRead"
```

### Advanced Examples

For more information about using template variables and defining keybindings, see the main [Keybindings][02] documentation.

<!-- Link reference definitions -->

[01]: ../_index.md
[02]: ../../getting-started/keybindings/_index.md
