package listviewport

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Model struct {
	ctx             context.ProgramContext
	viewport        viewport.Model
	topBoundId      int
	bottomBoundId   int
	currId          int
	ListItemHeight  int
	NumCurrentItems int
	NumTotalItems   int
	LastUpdated     time.Time
	CreatedAt       time.Time
	ItemTypeLabel   string
}

func NewModel(
	ctx context.ProgramContext,
	dimensions constants.Dimensions,
	lastUpdated time.Time,
	createdAt time.Time,
	itemTypeLabel string,
	numItems, listItemHeight int,
) Model {
	model := Model{
		ctx:             ctx,
		NumCurrentItems: numItems,
		ListItemHeight:  listItemHeight,
		currId:          0,
		viewport: viewport.Model{
			Width:  dimensions.Width,
			Height: dimensions.Height,
		},
		topBoundId:    0,
		ItemTypeLabel: itemTypeLabel,
		LastUpdated:   lastUpdated,
		CreatedAt:     createdAt,
	}
	model.bottomBoundId = utils.Min(
		model.NumCurrentItems-1,
		model.getNumPrsPerPage()-1,
	)
	return model
}

func (m *Model) SetNumItems(numItems int) {
	oldNumItems := m.NumCurrentItems
	m.NumCurrentItems = numItems
	itemsPerPage := m.getNumPrsPerPage()

	// Only reset bounds if this is initial setup (no items before) or items decreased
	if oldNumItems == 0 || numItems < oldNumItems {
		// Reset bounds for initial setup or when items are removed
		m.topBoundId = 0
		m.bottomBoundId = utils.Min(m.NumCurrentItems-1, itemsPerPage-1)
	} else {
		// Items were added (like page 2) - preserve current viewport position
		// Only extend bottomBoundId if current viewport can show more items
		maxVisibleBottom := m.topBoundId + itemsPerPage - 1
		m.bottomBoundId = utils.Min(m.NumCurrentItems-1, maxVisibleBottom)
	}
}

func (m *Model) SetTotalItems(total int) {
	m.NumTotalItems = total
}

func (m *Model) SetItemHeight(height int) {
	m.ListItemHeight = height
}

func (m *Model) SyncViewPort(content string) {
	m.viewport.SetContent(content)
}

func (m *Model) getNumPrsPerPage() int {
	if m.ListItemHeight == 0 {
		return 0
	}
	return m.viewport.Height / m.ListItemHeight
}

func (m *Model) ResetCurrItem() {
	m.currId = 0
	m.viewport.GotoTop()
}

func (m *Model) GetCurrItem() int {
	return m.currId
}

func (m *Model) SetCurrItem(item int) {
	if item < 0 {
		m.currId = 0
	} else if item >= m.NumCurrentItems {
		m.currId = m.NumCurrentItems - 1
	} else {
		m.currId = item
	}

	// Adjust viewport to keep current item visible
	itemsPerPage := m.getNumPrsPerPage()
	if m.currId < m.topBoundId {
		// Item is above visible area, scroll up
		m.topBoundId = m.currId
		m.bottomBoundId = utils.Min(m.NumCurrentItems-1, m.topBoundId+itemsPerPage-1)
		m.viewport.GotoTop()
		// Scroll down to position correctly
		scrollLines := m.currId * m.ListItemHeight
		for i := 0; i < scrollLines; i += m.ListItemHeight {
			m.viewport.ScrollDown(m.ListItemHeight)
		}
	} else if m.currId > m.bottomBoundId {
		// Item is below visible area, scroll down
		m.bottomBoundId = m.currId
		m.topBoundId = utils.Max(0, m.bottomBoundId-itemsPerPage+1)
		m.viewport.GotoTop()
		// Scroll down to position correctly
		scrollLines := m.topBoundId * m.ListItemHeight
		for i := 0; i < scrollLines; i += m.ListItemHeight {
			m.viewport.ScrollDown(m.ListItemHeight)
		}
	}
}

func (m *Model) GetVisibleBounds() (int, int) {
	return m.topBoundId, m.bottomBoundId
}

func (m *Model) NextItem() int {
	// Don't go beyond the last item
	if m.currId >= m.NumCurrentItems-1 {
		return m.currId
	}

	// Move cursor down first
	m.currId += 1

	// If cursor is now beyond the bottom of visible area, scroll viewport down
	if m.currId > m.bottomBoundId {
		m.topBoundId += 1
		m.bottomBoundId += 1
		m.viewport.ScrollDown(m.ListItemHeight)
	}

	return m.currId
}

func (m *Model) PrevItem() int {
	// Don't go beyond the first item
	if m.currId <= 0 {
		return m.currId
	}

	// Move cursor up first
	m.currId -= 1

	// If cursor is now above the top of visible area, scroll viewport up
	if m.currId < m.topBoundId {
		m.topBoundId -= 1
		m.bottomBoundId -= 1
		m.viewport.ScrollUp(m.ListItemHeight)
	}

	return m.currId
}

func (m *Model) FirstItem() int {
	m.currId = 0
	m.topBoundId = 0
	m.bottomBoundId = utils.Min(m.NumCurrentItems-1, m.getNumPrsPerPage()-1)
	m.viewport.GotoTop()
	return m.currId
}

func (m *Model) LastItem() int {
	m.currId = m.NumCurrentItems - 1
	// Update bounds to reflect bottom position
	itemsPerPage := m.getNumPrsPerPage()
	if m.NumCurrentItems > itemsPerPage {
		m.bottomBoundId = m.NumCurrentItems - 1
		m.topBoundId = m.NumCurrentItems - itemsPerPage
	} else {
		m.topBoundId = 0
		m.bottomBoundId = m.NumCurrentItems - 1
	}
	m.viewport.GotoBottom()
	return m.currId
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.viewport.Height = max(0, dimensions.Height)
	m.viewport.Width = max(0, dimensions.Width)
}

func (m *Model) View() string {
	viewport := m.viewport.View()
	return lipgloss.NewStyle().
		Width(m.viewport.Width).
		MaxWidth(m.viewport.Width).
		Render(
			viewport,
		)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = *ctx
}
