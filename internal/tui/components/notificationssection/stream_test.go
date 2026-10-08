package notificationssection

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func notif(id, owner string) data.NotificationData {
	n := data.NotificationData{Id: id, Unread: true}
	n.Repository.Owner.Login = owner
	return n
}

func TestSharedStreamSplitsPagesByOrg(t *testing.T) {
	restore := data.OverrideDoneStoreForTesting(
		data.NewDoneStoreForTesting(filepath.Join(t.TempDir(), "done.json")),
	)
	defer restore()

	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	ctx := &context.ProgramContext{Config: &cfg}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	newSection := func(filters string) *Model {
		m := NewModel(0, ctx, config.NotificationsSectionConfig{Filters: filters}, time.Now())
		return &m
	}
	all := newSection("is:all")
	org := newSection("is:all org:onomondo")
	other := newSection("is:all -org:onomondo")
	sections := []*Model{all, org, other}

	key := all.filters().streamKey()
	page1 := SectionNotificationsFetchedMsg{
		Notifications: []data.NotificationData{notif("1", "onomondo"), notif("2", "joelazar")},
		StreamKey:     key,
		PageInfo:      data.PageInfo{HasNextPage: true, EndCursor: "2"},
	}
	page2 := SectionNotificationsFetchedMsg{
		Notifications: []data.NotificationData{notif("3", "Onomondo")},
		StreamKey:     key,
		From:          "2",
		PageInfo:      data.PageInfo{EndCursor: "3"},
	}
	for _, msg := range []SectionNotificationsFetchedMsg{page1, page2, page2} {
		for _, s := range sections {
			s.Update(msg)
		}
	}

	for name, tc := range map[string]struct {
		m    *Model
		want int
	}{"all": {all, 3}, "org": {org, 2}, "other": {other, 1}} {
		if got := len(tc.m.Notifications); got != tc.want {
			t.Errorf("%s: got %d notifications, want %d", name, got, tc.want)
		}
	}
}
