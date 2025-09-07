---
title: Selected Notification
linkTitle: >-
  ![icon:bell](lucide)&nbsp;Selected Notification
summary: >-
  Key bindings for acting on a selected notification.
weight: 4
---

When you select a notification in the Notifications view, you can perform actions on it using
these key combinations.

| Keystroke        | Action          | Description                                                                  |
| :--------------- | :-------------- | :--------------------------------------------------------------------------- |
| ![kbd:`Enter`]() | Open            | Open the notification URL in your default web browser.                       |
| ![kbd:`m`]()     | Mark as Read    | Mark the selected notification as read on GitHub.                            |
| ![kbd:`d`]()     | Mark as Done    | Mark the selected notification as done (read and remove from notifications). |
| ![kbd:`u`]()     | Mark as Unread  | Mark the selected notification as unread on GitHub.                          |
| ![kbd:`b`]()     | Open in Browser | Open the notification URL in your default web browser.                       |
| ![kbd:`y`]()     | Copy URL        | Copy the notification URL to your clipboard.                                 |

## Custom Keybindings

You can define custom keybindings for notifications in your [configuration file][01] under the
`keybindings.notifications` setting.

```yaml
keybindings:
  notifications:
    - key: "ctrl+r"
      builtin: "markRead"
    - key: "ctrl+d"
      builtin: "markDone"
```

### Available Built-in Actions

| Action       | Description                    |
| :----------- | :----------------------------- |
| `markRead`   | Mark the notification as read  |
| `markDone`   | Mark the notification as done  |
| `viewSwitch` | Switch between different views |

### Template Variables

When defining custom commands for notifications, you can use these template variables:

| Variable                    | Description                                                   |
| :-------------------------- | :------------------------------------------------------------ |
| `{{.Title}}`                | The notification title                                        |
| `{{.GetRepoNameWithOwner}}` | The repository name with owner (e.g., `owner/repo`)           |
| `{{.URL}}`                  | The notification URL                                          |
| `{{.ThreadID}}`             | The notification thread ID                                    |
| `{{.Type}}`                 | The notification type (e.g., `PullRequest`, `Issue`)          |
| `{{.Reason}}`               | The notification reason (e.g., `review_requested`, `mention`) |

<!-- Link reference definitions -->

[01]: ../../configuration/_index.md
