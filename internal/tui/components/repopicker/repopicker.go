// Package repopicker is a fuzzy picker that switches the repo gh-dash is scoped to.
package repopicker

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sahilm/fuzzy"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

const (
	width      = 60
	maxVisible = 12
	allLabel   = "all repos (no repo filter)"
)

// Repo is a local clone. An empty Name means "no repo scope".
type Repo struct {
	Name string // owner/name
	Path string
}

func (r Repo) label() string {
	if r.Name == "" {
		return allLabel
	}
	return r.Name
}

// ListLocalRepos expands the repoPaths config into the local git clones it points at.
// "owner/*: ~/dir/*" lists every git repo in ~/dir; exact "owner/name: path" entries are kept as is.
func ListLocalRepos(repoPaths map[string]string) []Repo {
	seen := map[string]bool{}
	var repos []Repo
	add := func(name, path string) {
		if !seen[name] {
			seen[name] = true
			repos = append(repos, Repo{Name: name, Path: path})
		}
	}
	for key, val := range repoPaths {
		val = expandHome(val)
		owner, isWildcard := strings.CutSuffix(key, "/*")
		switch {
		case isWildcard && strings.HasSuffix(val, "/*"):
			dir := strings.TrimSuffix(val, "/*")
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				path := filepath.Join(dir, e.Name())
				if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
					add(owner+"/"+e.Name(), path)
				}
			}
		case !strings.ContainsAny(key, "*:") && strings.Count(key, "/") == 1:
			add(key, val)
		}
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return home + p[1:]
		}
	}
	return p
}

type labels []Repo

func (l labels) String(i int) string { return l[i].label() }
func (l labels) Len() int            { return len(l) }

type item struct {
	repo    Repo
	matched []int
}

type Model struct {
	ctx      *context.ProgramContext
	input    textinput.Model
	repos    []Repo
	filtered []item
	selected int
	open     bool
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{ctx: ctx}
}

func (m *Model) IsOpen() bool { return m.open }

func (m *Model) Open(repos []Repo) tea.Cmd {
	m.repos = append([]Repo{{}}, repos...)
	m.input = inputbox.DefaultTextInput(m.ctx)
	m.input.Placeholder = "type to filter repos"
	m.input.SetWidth(width - 22)
	m.open = true
	m.filter()
	return m.input.Focus()
}

func (m *Model) filter() {
	m.selected = 0
	m.filtered = nil
	q := m.input.Value()
	if q == "" {
		for _, r := range m.repos {
			m.filtered = append(m.filtered, item{repo: r})
		}
		return
	}
	for _, match := range fuzzy.FindFrom(q, labels(m.repos)) {
		m.filtered = append(m.filtered, item{m.repos[match.Index], match.MatchedIndexes})
	}
}

// Update handles a key while the picker is open. It returns the chosen repo on enter.
func (m *Model) Update(msg tea.KeyMsg) (tea.Cmd, *Repo) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.open = false
		return nil, nil
	case "enter":
		m.open = false
		if m.selected < len(m.filtered) {
			r := m.filtered[m.selected].repo
			return nil, &r
		}
		return nil, nil
	case "up", "ctrl+p", "ctrl+k":
		if m.selected > 0 {
			m.selected--
		}
		return nil, nil
	case "down", "ctrl+n", "ctrl+j":
		if m.selected < len(m.filtered)-1 {
			m.selected++
		}
		return nil, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filter()
	return cmd, nil
}

func (m *Model) View() string {
	if !m.open {
		return ""
	}
	s := m.ctx.Styles.Select
	faint := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)
	divider := faint.Render(strings.Repeat("─", width))

	title := lipgloss.NewStyle().Bold(true).Foreground(m.ctx.Theme.PrimaryText).Render("Switch repo")
	count := faint.Render(fmt.Sprintf("%d/%d", len(m.filtered), len(m.repos)))
	header := title + "  " + m.input.View()
	header += strings.Repeat(" ", max(1, width-lipgloss.Width(header)-lipgloss.Width(count))) + count

	current := ""
	if m.ctx.HasGHRepo() {
		current = m.ctx.GHRepo.Owner + "/" + m.ctx.GHRepo.Name
	}

	// Scroll so the selected row stays in view.
	start := max(0, m.selected-maxVisible+1)
	end := min(len(m.filtered), start+maxVisible)

	lines := []string{header, divider}
	if len(m.filtered) == 0 {
		lines = append(lines, s.NoResults.Render("no matching repos"))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.row(m.filtered[i], i == m.selected, m.filtered[i].repo.Name == current))
	}
	// Fixed height so the popup does not jump while filtering.
	for len(lines) < min(maxVisible, len(m.repos))+2 {
		lines = append(lines, "")
	}

	key := lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText)
	help := key.Render("enter") + faint.Render(" select   ") +
		key.Render("esc") + faint.Render(" cancel   ") +
		key.Render("↑/↓") + faint.Render(" move")
	lines = append(lines, divider, help)

	return s.PopupStyle.Padding(0, 1).Width(width + 4).Render(strings.Join(lines, "\n"))
}

func (m *Model) row(it item, selected, current bool) string {
	t := m.ctx.Theme
	base := lipgloss.NewStyle()
	if selected {
		base = base.Background(t.SelectedBackground)
	}
	owner := base.Foreground(t.FaintText)
	name := base.Foreground(t.SecondaryText)
	if selected {
		name = base.Foreground(t.PrimaryText).Bold(true)
	}
	if it.repo.Name == "" {
		name = owner.Italic(true)
	}
	hit := base.Foreground(m.ctx.Styles.Colors.OpenIssue).Bold(true)

	var b strings.Builder
	if selected {
		b.WriteString(m.ctx.Styles.Select.Pointer.Inherit(base).Render(constants.SelectionIcon + " "))
	} else {
		b.WriteString("  ")
	}
	label := it.repo.label()
	nameStart := strings.IndexByte(label, '/') + 1
	for i, r := range label {
		st := name
		if i < nameStart {
			st = owner
		}
		if slices.Contains(it.matched, i) {
			st = hit
		}
		b.WriteString(st.Render(string(r)))
	}

	suffix := ""
	if current {
		suffix = base.Foreground(t.FaintText).Render("current ")
	}
	gap := max(1, width-lipgloss.Width(b.String())-lipgloss.Width(suffix))
	return b.String() + base.Render(strings.Repeat(" ", gap)) + suffix
}
