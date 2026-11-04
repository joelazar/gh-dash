// Package repopicker is a fuzzy picker that switches the repo gh-dash is scoped to.
package repopicker

import (
	"os"
	"path/filepath"
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

type Model struct {
	ctx      *context.ProgramContext
	input    textinput.Model
	repos    []Repo
	filtered []Repo
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
	m.input.SetWidth(width - 14)
	m.open = true
	m.filter()
	return m.input.Focus()
}

func (m *Model) filter() {
	m.selected = 0
	q := m.input.Value()
	if q == "" {
		m.filtered = m.repos
		return
	}
	m.filtered = nil
	for _, match := range fuzzy.FindFrom(q, labels(m.repos)) {
		m.filtered = append(m.filtered, m.repos[match.Index])
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
			r := m.filtered[m.selected]
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
	current := ""
	if m.ctx.HasGHRepo() {
		current = m.ctx.GHRepo.Owner + "/" + m.ctx.GHRepo.Name
	}
	faint := lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText)

	// Scroll so the selected row stays in view.
	start := max(0, m.selected-maxVisible+1)
	end := min(len(m.filtered), start+maxVisible)

	var b strings.Builder
	b.WriteString(" Switch repo: " + m.input.View() + "\n")
	b.WriteString(faint.Render(strings.Repeat("─", width)) + "\n")
	if len(m.filtered) == 0 {
		b.WriteString(s.NoResults.Render("no matching repos") + "\n")
	}
	for i := start; i < end; i++ {
		r := m.filtered[i]
		prefix := "  "
		style := s.ItemStyle.Foreground(m.ctx.Theme.SecondaryText)
		if i == m.selected {
			prefix = s.Pointer.Render(constants.SelectionIcon + " ")
			style = s.SelectedStyle.Bold(true)
		}
		line := prefix + style.Render(r.label())
		if r.Name == current {
			line += faint.Render(" (current)")
		}
		b.WriteString(line + "\n")
	}
	b.WriteString(faint.Render(" enter select · esc cancel · ↑/↓ move"))
	return s.PopupStyle.Width(width + 2).Render(b.String())
}
