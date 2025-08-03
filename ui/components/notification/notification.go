package notification

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/components"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type Notification struct {
	Ctx  *context.ProgramContext
	Data *data.Notification
}

func (n *Notification) getTextStyle() lipgloss.Style {
	return components.GetIssueTextStyle(n.Ctx)
}

func (n *Notification) renderRepository() string {
	return n.getTextStyle().
		Foreground(n.Ctx.Theme.FaintText).
		Render(n.Data.Repository)
}

func (n *Notification) renderTitle() string {
	return n.getTextStyle().
		Render(n.Data.Title)
}

func (n *Notification) renderType() string {
	notificationType := n.Data.Type
	if notificationType == "PullRequest" {
		notificationType = "PR"
	}

	return n.getTextStyle().
		Foreground(n.Ctx.Theme.SecondaryText).
		Render(notificationType)
}

func (n *Notification) renderReason() string {
	return n.getTextStyle().
		Foreground(n.Ctx.Theme.SecondaryText).
		Render(n.Data.Reason.Format())
}

func (n *Notification) renderUnreadIcon() string {
	iconStyle := lipgloss.NewStyle()

	if n.Data.Unread {
		iconStyle = iconStyle.Foreground(n.Ctx.Theme.WarningText)
		return iconStyle.Render("●")
	}

	iconStyle = iconStyle.Foreground(n.Ctx.Theme.FaintText)
	return iconStyle.Render("○")
}

func (n *Notification) renderUpdatedAt() string {
	return n.getTextStyle().
		Foreground(n.Ctx.Theme.FaintText).
		Render(utils.TimeElapsed(n.Data.UpdatedAt))
}

func (n *Notification) ToTableRow() table.Row {
	row := table.Row{
		n.renderRepository(),
		n.renderTitle(),
		n.renderType(),
		n.renderReason(),
		n.renderUnreadIcon(),
		n.renderUpdatedAt(),
	}

	return row
}
