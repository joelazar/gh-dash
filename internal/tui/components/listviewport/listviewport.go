package listviewport

import (
	"time"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type Model struct {
	ctx             context.ProgramContext
	viewport        viewport.Model
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
	return Model{
		ctx:             ctx,
		NumCurrentItems: numItems,
		ListItemHeight:  listItemHeight,
		currId:          0,
		viewport: viewport.New(
			viewport.WithWidth(dimensions.Width),
			viewport.WithHeight(dimensions.Height),
		),
		ItemTypeLabel: itemTypeLabel,
		LastUpdated:   lastUpdated,
		CreatedAt:     createdAt,
	}
}

func (m *Model) SetNumItems(numItems int) {
	m.NumCurrentItems = numItems
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

func (m *Model) ResetCurrItem() {
	m.currId = 0
	m.viewport.GotoTop()
}

func (m *Model) GetCurrItem() int {
	return m.currId
}

func (m *Model) NextItem() int {
	m.currId = utils.Max(utils.Min(m.currId+1, m.NumCurrentItems-1), 0)
	m.scrollToCurrItem()
	return m.currId
}

func (m *Model) PrevItem() int {
	m.currId = utils.Max(m.currId-1, 0)
	m.scrollToCurrItem()
	return m.currId
}

func (m *Model) scrollToCurrItem() {
	top := m.currId * m.ListItemHeight
	bottom := top + m.ListItemHeight
	if top < m.viewport.YOffset() {
		m.viewport.SetYOffset(top)
	} else if bottom > m.viewport.YOffset()+m.viewport.Height() {
		m.viewport.SetYOffset(bottom - m.viewport.Height())
	}
}

func (m *Model) FirstItem() int {
	m.currId = 0
	m.viewport.GotoTop()
	return m.currId
}

func (m *Model) LastItem() int {
	m.currId = m.NumCurrentItems - 1
	m.viewport.GotoBottom()
	return m.currId
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.viewport.SetHeight(max(0, dimensions.Height))
	m.viewport.SetWidth(max(0, dimensions.Width))
}

func (m *Model) View() string {
	viewport := m.viewport.View()
	return lipgloss.NewStyle().
		Width(m.viewport.Width()).
		MaxWidth(m.viewport.Width()).
		Render(
			viewport,
		)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = *ctx
}
