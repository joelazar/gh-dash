package listviewport

import (
	"strings"
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func newTestModel(numItems, viewportHeight int) Model {
	dims := constants.Dimensions{Width: 80, Height: viewportHeight}
	now := time.Now()
	m := NewModel(context.ProgramContext{}, dims, now, now, "items", 0, 1)
	setItems(&m, numItems)
	return m
}

func setItems(m *Model, n int) {
	m.SetNumItems(n)
	m.SyncViewPort(strings.Repeat("row\n", n-1) + "row")
}

func assertVisible(t *testing.T, m Model) {
	t.Helper()
	top := m.viewport.YOffset()
	if m.currId < top || m.currId >= top+m.viewport.Height() {
		t.Fatalf("item %d not visible in rows %d-%d", m.currId, top, top+m.viewport.Height()-1)
	}
}

func TestNextItemScrollsOnlyPastBottom(t *testing.T) {
	m := newTestModel(10, 5)
	for range 4 {
		m.NextItem()
	}
	if m.viewport.YOffset() != 0 {
		t.Fatalf("scrolled before reaching the bottom, offset %d", m.viewport.YOffset())
	}
	m.NextItem()
	if m.viewport.YOffset() != 1 {
		t.Fatalf("expected offset 1, got %d", m.viewport.YOffset())
	}
	assertVisible(t, m)
}

func TestPrevItemScrollsOnlyPastTop(t *testing.T) {
	m := newTestModel(10, 5)
	for range 5 {
		m.NextItem()
	}
	for range 4 {
		m.PrevItem()
	}
	if m.viewport.YOffset() != 1 {
		t.Fatalf("scrolled before reaching the top, offset %d", m.viewport.YOffset())
	}
	m.PrevItem()
	if m.GetCurrItem() != 0 || m.viewport.YOffset() != 0 {
		t.Fatalf("expected item 0 at offset 0, got item %d offset %d", m.GetCurrItem(), m.viewport.YOffset())
	}
}

func TestPrevItemAfterLastItemKeepsCursorVisible(t *testing.T) {
	m := newTestModel(40, 5)
	m.LastItem()
	setItems(&m, 80)
	m.LastItem()
	for range 20 {
		m.PrevItem()
		assertVisible(t, m)
	}
	for range 20 {
		m.NextItem()
		assertVisible(t, m)
	}
	if m.GetCurrItem() != 79 {
		t.Fatalf("expected item 79, got %d", m.GetCurrItem())
	}
}

func TestPrevItemAtFirstItem(t *testing.T) {
	m := newTestModel(10, 5)
	m.PrevItem()
	if m.GetCurrItem() != 0 || m.viewport.YOffset() != 0 {
		t.Errorf("expected item 0 at offset 0, got item %d offset %d", m.GetCurrItem(), m.viewport.YOffset())
	}
}

func TestNextItemAtLastItem(t *testing.T) {
	m := newTestModel(10, 5)
	for range 20 {
		m.NextItem()
	}
	if m.GetCurrItem() != 9 {
		t.Errorf("expected currId=9, got %d", m.GetCurrItem())
	}
	assertVisible(t, m)
}
