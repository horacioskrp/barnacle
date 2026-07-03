package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sessionState represents the current screen displayed by the TUI.
type sessionState int

const (
	stateLoading sessionState = iota
	stateBrowsing
	stateConfirming
	statePruning
	stateSummary
	stateError
)

const gaugeWidth = 40

// Styles

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("31")).
			Padding(0, 2).
			MarginBottom(1)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	staleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	gaugeFilledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39"))

	gaugeEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("31")).
			Padding(1, 2)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	warningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1, 2)
)

// Messages produced by asynchronous commands.

type diskUsageMsg struct {
	categories []Category
}

type pruneResultMsg struct {
	summary PruneSummary
}

type errMsg struct {
	err error
}

// model is the single source of truth for Barnacle's TUI (MVU pattern).
type model struct {
	docker     *DockerClient
	state      sessionState
	categories []Category
	cursor     int
	selected   map[CategoryID]bool
	summary    PruneSummary
	err        error
}

// initialModel builds the starting state, ready to load disk usage.
func initialModel(docker *DockerClient) model {
	return model{
		docker:   docker,
		state:    stateLoading,
		selected: make(map[CategoryID]bool),
	}
}

// Init kicks off the initial disk-usage analysis.
func (m model) Init() tea.Cmd {
	return loadDiskUsageCmd(m.docker)
}

func loadDiskUsageCmd(docker *DockerClient) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		categories, err := docker.Analyze(ctx)
		if err != nil {
			return errMsg{err: err}
		}
		return diskUsageMsg{categories: categories}
	}
}

func runPruneCmd(docker *DockerClient, selected map[CategoryID]bool) tea.Cmd {
	chosen := make(map[CategoryID]bool, len(selected))
	for id, ok := range selected {
		chosen[id] = ok
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		summary := docker.Prune(ctx, chosen)
		return pruneResultMsg{summary: summary}
	}
}

// Update handles incoming messages and advances the model accordingly.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diskUsageMsg:
		m.categories = msg.categories
		m.state = stateBrowsing
		return m, nil

	case pruneResultMsg:
		m.summary = msg.summary
		m.state = stateSummary
		return m, nil

	case errMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if (key == "ctrl+c" || key == "q") && m.state != statePruning {
		return m, tea.Quit
	}

	switch m.state {
	case stateBrowsing:
		return m.handleBrowsingKey(key)
	case stateConfirming:
		return m.handleConfirmingKey(key)
	case stateSummary, stateError:
		if key == "enter" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) handleBrowsingKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.categories)-1 {
			m.cursor++
		}

	case " ":
		if len(m.categories) > 0 {
			id := m.categories[m.cursor].ID
			m.selected[id] = !m.selected[id]
		}

	case "enter":
		if m.anySelected() {
			m.state = stateConfirming
		}
	}

	return m, nil
}

func (m model) handleConfirmingKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "enter":
		m.state = statePruning
		return m, runPruneCmd(m.docker, m.selected)

	case "n", "esc":
		m.state = stateBrowsing
	}

	return m, nil
}

func (m model) anySelected() bool {
	for _, ok := range m.selected {
		if ok {
			return true
		}
	}
	return false
}

// View renders the current screen.
func (m model) View() string {
	switch m.state {
	case stateLoading:
		return m.viewLoading()
	case stateConfirming:
		return m.viewConfirm()
	case statePruning:
		return m.viewPruning()
	case stateSummary:
		return m.viewSummary()
	case stateError:
		return m.viewError()
	default:
		return m.viewBrowsing()
	}
}

func header() string {
	return titleStyle.Render("🐋 BARNACLE — nettoyage Docker")
}

func (m model) viewLoading() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("⏳ Analyse du démon Docker en cours..."))
	b.WriteString("\n")
	return b.String()
}

func (m model) viewPruning() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("🧹 Grattage des bernacles en cours, veuillez patienter..."))
	b.WriteString("\n")
	return b.String()
}

func (m model) viewError() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")
	b.WriteString(boxStyle.Render(errorStyle.Render("✗ " + m.err.Error())))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("entrée/q : quitter"))
	b.WriteString("\n")
	return b.String()
}

func (m model) totalSize() int64 {
	var total int64
	for _, cat := range m.categories {
		total += cat.Size
	}
	return total
}

func (m model) selectedSize() int64 {
	var total int64
	for _, cat := range m.categories {
		if m.selected[cat.ID] {
			total += cat.Size
		}
	}
	return total
}

func (m model) viewBrowsing() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")

	total := m.totalSize()
	selected := m.selectedSize()
	b.WriteString(renderGauge(selected, total))
	b.WriteString("\n\n")

	for i, cat := range m.categories {
		b.WriteString(renderCategoryRow(cat, i == m.cursor, m.selected[cat.ID]))
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render(
		"↑/k ↓/j : naviguer  •  espace : sélectionner  •  entrée : confirmer la sélection  •  q : quitter",
	))
	b.WriteString("\n")

	return b.String()
}

func renderGauge(selected, total int64) string {
	label := fmt.Sprintf("Espace récupérable : %s / %s sélectionnés", FormatBytes(selected), FormatBytes(total))

	filled := 0
	if total > 0 {
		filled = int(float64(gaugeWidth) * float64(selected) / float64(total))
		if filled > gaugeWidth {
			filled = gaugeWidth
		}
	}
	empty := gaugeWidth - filled

	bar := gaugeFilledStyle.Render(strings.Repeat("█", filled)) +
		gaugeEmptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s\n[%s]", label, bar)
}

func renderCategoryRow(cat Category, isCursor, isSelected bool) string {
	cursor := "  "
	if isCursor {
		cursor = cursorStyle.Render("➤ ")
	}

	checkbox := "[ ]"
	if isSelected {
		checkbox = selectedStyle.Render("[x]")
	}

	label := fmt.Sprintf("%-26s %10s", cat.Label, FormatBytes(cat.Size))
	if cat.Count > 0 {
		label += fmt.Sprintf("  (%d élément%s)", cat.Count, plural(cat.Count))
	}

	if isSelected {
		label = selectedStyle.Render(label)
	}

	row := fmt.Sprintf("%s%s %s", cursor, checkbox, label)

	if cat.HasStale {
		row += "  " + staleStyle.Render(fmt.Sprintf("⚠ inutilisé depuis %s", formatAge(cat.OldestAge)))
	}

	return row
}

func (m model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")
	b.WriteString(confirmTitleStyle.Render("🧐 Confirmer le nettoyage ?"))
	b.WriteString("\n\n")

	var total int64
	hasVolumes := false

	for _, cat := range m.categories {
		if !m.selected[cat.ID] {
			continue
		}
		total += cat.Size
		if cat.ID == CategoryOrphanVolumes {
			hasVolumes = true
		}

		line := fmt.Sprintf("  • %-26s %10s", cat.Label, FormatBytes(cat.Size))
		if cat.Count > 0 {
			line += fmt.Sprintf("  (%d élément%s)", cat.Count, plural(cat.Count))
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Espace total qui sera libéré : %s", FormatBytes(total)))
	b.WriteString("\n\n")

	if hasVolumes {
		b.WriteString(warningBoxStyle.Render(
			"⚠ Des volumes sont sélectionnés.\nIls peuvent contenir des données importantes.\nCette action est IRRÉVERSIBLE.",
		))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render(
		"y/entrée : confirmer  •  n/échap : annuler et revenir à la sélection  •  q : quitter",
	))
	b.WriteString("\n")

	return b.String()
}

func plural(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func formatAge(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days < 1 {
		return "moins d'1 jour"
	}
	if days == 1 {
		return "1 jour"
	}
	return fmt.Sprintf("%d jours", days)
}

func (m model) viewSummary() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n")
	b.WriteString(successStyle.Render("✓ Nettoyage terminé"))
	b.WriteString("\n\n")

	for _, res := range m.summary.Results {
		if res.Err != nil {
			b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s : %v", res.Label, res.Err)))
		} else {
			b.WriteString(fmt.Sprintf("✓ %-26s %s libérés", res.Label, FormatBytes(int64(res.SpaceReclaimed))))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(successStyle.Render(fmt.Sprintf("Total libéré : %s", FormatBytes(int64(m.summary.TotalReclaimed())))))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("entrée/q : quitter"))
	b.WriteString("\n")

	return b.String()
}
