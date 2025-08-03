package table

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/ui/common"
	"github.com/dlvhdr/gh-dash/v4/ui/components/listviewport"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type Model struct {
	ctx            context.ProgramContext
	Columns        []Column
	Rows           []Row
	EmptyState     *string
	loadingMessage string
	isLoading      bool
	loadingSpinner spinner.Model
	dimensions     constants.Dimensions
	rowsViewport   listviewport.Model
	// Performance optimization: cache rendered content
	cachedContent      string
	cachedContentValid bool
	lastSelectedRow    int
}

type Column struct {
	Title         string
	Hidden        *bool
	Width         *int
	ComputedWidth int
	Grow          *bool
}

type Row []string

func NewModel(
	ctx context.ProgramContext,
	dimensions constants.Dimensions,
	lastUpdated time.Time,
	createdAt time.Time,
	columns []Column,
	rows []Row,
	itemTypeLabel string,
	emptyState *string,
	loadingMessage string,
	isLoading bool,
) Model {
	itemHeight := 1
	if !ctx.Config.Theme.Ui.Table.Compact {
		itemHeight += 1
	}
	if ctx.Config.Theme.Ui.Table.ShowSeparator {
		itemHeight += 1
	}

	loadingSpinner := spinner.New()
	loadingSpinner.Spinner = spinner.Dot
	loadingSpinner.Style = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText)

	return Model{
		ctx:            ctx,
		Columns:        columns,
		Rows:           rows,
		EmptyState:     emptyState,
		loadingMessage: loadingMessage,
		isLoading:      isLoading,
		loadingSpinner: loadingSpinner,
		dimensions:     dimensions,
		rowsViewport: listviewport.NewModel(
			ctx,
			dimensions,
			lastUpdated,
			createdAt,
			itemTypeLabel,
			len(rows),
			itemHeight,
		),
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.isLoading {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	}
	return m, cmd
}

func (m Model) StartLoadingSpinner() tea.Cmd {
	return m.loadingSpinner.Tick
}

func (m Model) View() string {
	header := m.renderHeader()
	body := m.renderBody()

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *Model) SetIsLoading(isLoading bool) {
	m.isLoading = isLoading
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.dimensions = dimensions
	m.rowsViewport.SetDimensions(constants.Dimensions{
		Width:  m.dimensions.Width,
		Height: m.dimensions.Height,
	})
	// Invalidate cache when dimensions change
	m.cachedContentValid = false
}

func (m *Model) ResetCurrItem() {
	m.rowsViewport.ResetCurrItem()
}

func (m *Model) GetCurrItem() int {
	return m.rowsViewport.GetCurrItem()
}

func (m *Model) PrevItem() int {
	currItem := m.rowsViewport.PrevItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) NextItem() int {
	currItem := m.rowsViewport.NextItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) FirstItem() int {
	currItem := m.rowsViewport.FirstItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) LastItem() int {
	currItem := m.rowsViewport.LastItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) cacheColumnWidths() {
	columns := m.renderHeaderColumns()
	for i, col := range columns {
		if m.Columns[i].Hidden != nil && *m.Columns[i].Hidden {
			continue
		}
		m.Columns[i].ComputedWidth = lipgloss.Width(col)
	}
}

func (m *Model) createEmptyRow() string {
	// Create an empty row with the same width as regular rows but minimal content
	// This acts as a placeholder for virtual scrolling
	var rowParts []string
	
	for _, col := range m.getShownColumns() {
		width := col.ComputedWidth
		if width == 0 {
			width = 10 // fallback width
		}
		// Create empty space with same width as actual content
		emptyContent := strings.Repeat(" ", width)
		rowParts = append(rowParts, emptyContent)
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Left, rowParts...)
}

func (m *Model) SyncViewPortContent() {
	totalRows := len(m.Rows)
	if totalRows == 0 {
		m.rowsViewport.SyncViewPort("")
		m.cachedContentValid = false
		return
	}
	
	currentSelectedRow := m.rowsViewport.GetCurrItem()
	
	// Check if we can use cached content (no data changes, only selection changed)
	if m.cachedContentValid && currentSelectedRow != m.lastSelectedRow && totalRows > 50 {
		log.Debug("PERF: SyncViewPortContent using selection-only update", "totalRows", totalRows, "oldSelection", m.lastSelectedRow, "newSelection", currentSelectedRow)
		
		// For large lists, try to avoid full re-render when only selection changes
		// Re-render just the affected rows: old selection and new selection
		headerColumns := m.renderHeaderColumns()
		m.cacheColumnWidths()
		
		start := time.Now()
		
		// Split cached content into lines and update only the changed rows
		lines := strings.Split(m.cachedContent, "\n")
		if m.lastSelectedRow < len(lines) && currentSelectedRow < len(lines) {
			// Update old selected row (unselect it)
			if m.lastSelectedRow >= 0 && m.lastSelectedRow < totalRows {
				lines[m.lastSelectedRow] = m.renderRow(m.lastSelectedRow, headerColumns)
			}
			// Update new selected row (select it)
			if currentSelectedRow >= 0 && currentSelectedRow < totalRows {
				lines[currentSelectedRow] = m.renderRow(currentSelectedRow, headerColumns)
			}
			
			updatedContent := strings.Join(lines, "\n")
			m.rowsViewport.SyncViewPort(updatedContent)
			m.cachedContent = updatedContent
			m.lastSelectedRow = currentSelectedRow
			
			log.Debug("PERF: SyncViewPortContent selection update complete", "totalRows", totalRows, "updateTime", time.Since(start))
			return
		}
	}
	
	// Full re-render needed (data changed or cache invalid)
	if totalRows > 100 {
		log.Debug("PERF: SyncViewPortContent full render", "totalRows", totalRows, "reason", "data_changed_or_cache_invalid")
	}
	
	headerColumns := m.renderHeaderColumns()
	m.cacheColumnWidths()
	
	start := time.Now()
	renderedRows := make([]string, 0, totalRows)
	for i := range m.Rows {
		renderedRows = append(renderedRows, m.renderRow(i, headerColumns))
	}
	renderTime := time.Since(start)
	
	content := lipgloss.JoinVertical(lipgloss.Left, renderedRows...)
	
	syncStart := time.Now()
	m.rowsViewport.SyncViewPort(content)
	syncTime := time.Since(syncStart)
	
	// Cache the rendered content
	m.cachedContent = content
	m.cachedContentValid = true
	m.lastSelectedRow = currentSelectedRow
	
	if totalRows > 100 {
		log.Debug("PERF: SyncViewPortContent full render complete", "totalRows", totalRows, "renderTime", renderTime, "syncTime", syncTime, "totalTime", time.Since(start))
	}
}

func (m *Model) SetRows(rows []Row) {
	m.Rows = rows
	m.rowsViewport.SetNumItems(len(m.Rows))
	// Invalidate cache when data changes
	m.cachedContentValid = false
	m.SyncViewPortContent()
}

func (m *Model) OnLineDown() {
	m.rowsViewport.NextItem()
}

func (m *Model) OnLineUp() {
	m.rowsViewport.PrevItem()
}

func (m *Model) getShownColumns() []Column {
	shownColumns := make([]Column, 0, len(m.Columns))
	for _, col := range m.Columns {
		if col.Hidden != nil && *col.Hidden {
			continue
		}

		shownColumns = append(shownColumns, col)
	}
	return shownColumns
}

func (m *Model) renderHeaderColumns() []string {
	shownColumns := m.getShownColumns()
	renderedColumns := make([]string, len(shownColumns))
	takenWidth := 0
	numGrowingColumns := 0
	for i, column := range shownColumns {
		if column.Grow != nil && *column.Grow {
			numGrowingColumns += 1
			continue
		}

		if column.Width != nil {
			renderedColumns[i] = m.ctx.Styles.Table.TitleCellStyle.
				Width(*column.Width).
				MaxWidth(*column.Width).
				Render(column.Title)
			takenWidth += *column.Width
			continue
		}

		cell := m.ctx.Styles.Table.TitleCellStyle.Render(column.Title)
		renderedColumns[i] = cell
		takenWidth += lipgloss.Width(cell)
	}

	if numGrowingColumns == 0 {
		return renderedColumns
	}

	leftoverWidth := m.dimensions.Width - takenWidth
	growCellWidth := leftoverWidth / numGrowingColumns
	for i, column := range shownColumns {
		if column.Grow == nil || !*column.Grow {
			continue
		}

		renderedColumns[i] = m.ctx.Styles.Table.TitleCellStyle.
			Width(growCellWidth).
			MaxWidth(growCellWidth).
			Render(column.Title)
	}

	return renderedColumns
}

func (m *Model) renderHeader() string {
	headerColumns := m.renderHeaderColumns()
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerColumns...)
	return m.ctx.Styles.Table.HeaderStyle.
		Width(m.dimensions.Width).
		MaxWidth(m.dimensions.Width).
		Height(common.TableHeaderHeight).
		MaxHeight(common.TableHeaderHeight).
		Render(header)
}

func (m *Model) renderBody() string {
	bodyStyle := lipgloss.NewStyle().
		Height(m.dimensions.Height).
		MaxWidth(m.dimensions.Width)

	if m.isLoading {
		return lipgloss.Place(
			m.dimensions.Width,
			m.dimensions.Height,
			lipgloss.Center,
			lipgloss.Center,
			fmt.Sprintf("%s%s", m.loadingSpinner.View(), m.loadingMessage),
		)
	}

	if len(m.Rows) == 0 && m.EmptyState != nil {
		return bodyStyle.Render(*m.EmptyState)
	}

	return m.rowsViewport.View()
}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	var style lipgloss.Style

	if m.rowsViewport.GetCurrItem() == rowId {
		style = m.ctx.Styles.Table.SelectedCellStyle
	} else {
		style = m.ctx.Styles.Table.CellStyle
	}

	renderedColumns := make([]string, 0, len(m.Columns))
	headerColId := 0

	for i, column := range m.Columns {
		if column.Hidden != nil && *column.Hidden {
			continue
		}

		colWidth := lipgloss.Width(headerColumns[headerColId])
		colHeight := 1
		if !m.ctx.Config.Theme.Ui.Table.Compact {
			colHeight = 2
		}
		col := m.Rows[rowId][i]
		renderedCol := style.
			Width(colWidth).
			MaxWidth(colWidth).
			Height(colHeight).
			MaxHeight(colHeight).
			Render(col)

		renderedColumns = append(renderedColumns, renderedCol)
		headerColId++
	}

	return m.ctx.Styles.Table.RowStyle.
		BorderBottom(m.ctx.Config.Theme.Ui.Table.ShowSeparator).
		MaxWidth(m.dimensions.Width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...))
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = *ctx
	m.rowsViewport.UpdateProgramContext(ctx)
}

func (m *Model) LastUpdated() time.Time {
	return m.rowsViewport.LastUpdated
}

func (m *Model) CreatedAt() time.Time {
	return m.rowsViewport.CreatedAt
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.rowsViewport.LastUpdated = t
}

func (m *Model) UpdateTotalItemsCount(count int) {
	m.rowsViewport.SetTotalItems(count)
}

func (m *Model) IsLoading() bool {
	return m.isLoading
}
