package main

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func key(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: t})
}

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}})
}

func testCategories() []Category {
	return []Category{
		{ID: CategoryDanglingImages, Label: "Images suspendues", Size: 100},
		{ID: CategoryStoppedContainers, Label: "Conteneurs arrêtés", Size: 200},
		{ID: CategoryOrphanVolumes, Label: "Volumes orphelins", Size: 300},
		{ID: CategoryBuildCache, Label: "Cache de build obsolète", Size: 400},
	}
}

func newBrowsingModel() model {
	m := initialModel(nil)
	m.state = stateBrowsing
	m.categories = testCategories()
	return m
}

func TestHandleBrowsingKeyNavigation(t *testing.T) {
	m := newBrowsingModel()

	m, _ = update(m, key(tea.KeyDown))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after moving down", m.cursor)
	}

	m, _ = update(m, key(tea.KeyUp))
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 after moving back up", m.cursor)
	}

	// Cannot go above the first item.
	m, _ = update(m, key(tea.KeyUp))
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (should not go negative)", m.cursor)
	}

	// Cannot go past the last item.
	for i := 0; i < 10; i++ {
		m, _ = update(m, runeKey('j'))
	}
	if m.cursor != len(m.categories)-1 {
		t.Fatalf("cursor = %d, want %d (should clamp to last item)", m.cursor, len(m.categories)-1)
	}
}

func TestHandleBrowsingKeyToggleSelection(t *testing.T) {
	m := newBrowsingModel()

	m, _ = update(m, key(tea.KeySpace))
	id := m.categories[m.cursor].ID
	if !m.selected[id] {
		t.Fatalf("category %v should be selected after pressing space", id)
	}

	m, _ = update(m, key(tea.KeySpace))
	if m.selected[id] {
		t.Fatalf("category %v should be unselected after pressing space again", id)
	}
}

func TestEnterWithoutSelectionStaysOnBrowsing(t *testing.T) {
	m := newBrowsingModel()

	m, cmd := update(m, key(tea.KeyEnter))
	if m.state != stateBrowsing {
		t.Fatalf("state = %v, want stateBrowsing when nothing is selected", m.state)
	}
	if cmd != nil {
		t.Fatalf("expected no command when nothing is selected")
	}
}

func TestEnterWithSelectionGoesToConfirming(t *testing.T) {
	m := newBrowsingModel()
	m.selected[CategoryDanglingImages] = true

	m, cmd := update(m, key(tea.KeyEnter))
	if m.state != stateConfirming {
		t.Fatalf("state = %v, want stateConfirming", m.state)
	}
	if cmd != nil {
		t.Fatalf("expected no command yet, pruning must wait for explicit confirmation")
	}
}

func TestConfirmingYesTriggersPrune(t *testing.T) {
	m := newBrowsingModel()
	m.state = stateConfirming
	m.selected[CategoryDanglingImages] = true

	m, cmd := update(m, runeKey('y'))
	if m.state != statePruning {
		t.Fatalf("state = %v, want statePruning", m.state)
	}
	if cmd == nil {
		t.Fatalf("expected a prune command to be returned")
	}
}

func TestConfirmingEnterTriggersPrune(t *testing.T) {
	m := newBrowsingModel()
	m.state = stateConfirming
	m.selected[CategoryDanglingImages] = true

	m, cmd := update(m, key(tea.KeyEnter))
	if m.state != statePruning {
		t.Fatalf("state = %v, want statePruning", m.state)
	}
	if cmd == nil {
		t.Fatalf("expected a prune command to be returned")
	}
}

func TestConfirmingNoGoesBackToBrowsingAndKeepsSelection(t *testing.T) {
	m := newBrowsingModel()
	m.state = stateConfirming
	m.selected[CategoryDanglingImages] = true

	m, cmd := update(m, runeKey('n'))
	if m.state != stateBrowsing {
		t.Fatalf("state = %v, want stateBrowsing", m.state)
	}
	if cmd != nil {
		t.Fatalf("expected no command when cancelling")
	}
	if !m.selected[CategoryDanglingImages] {
		t.Fatalf("selection should be preserved when cancelling back to browsing")
	}
}

func TestConfirmingEscGoesBackToBrowsing(t *testing.T) {
	m := newBrowsingModel()
	m.state = stateConfirming

	m, _ = update(m, key(tea.KeyEsc))
	if m.state != stateBrowsing {
		t.Fatalf("state = %v, want stateBrowsing", m.state)
	}
}

func TestQuitDisabledWhilePruning(t *testing.T) {
	m := newBrowsingModel()
	m.state = statePruning

	_, cmd := update(m, runeKey('q'))
	if cmd != nil {
		t.Fatalf("expected quit to be disabled while pruning is in progress")
	}
}

func TestQuitAllowedWhileBrowsing(t *testing.T) {
	m := newBrowsingModel()

	_, cmd := update(m, runeKey('q'))
	if cmd == nil {
		t.Fatalf("expected a quit command while browsing")
	}
}

func TestUpdateHandlesDiskUsageMsg(t *testing.T) {
	m := initialModel(nil)
	cats := testCategories()

	updated, _ := m.Update(diskUsageMsg{categories: cats})
	next := updated.(model)

	if next.state != stateBrowsing {
		t.Fatalf("state = %v, want stateBrowsing after disk usage loads", next.state)
	}
	if len(next.categories) != len(cats) {
		t.Fatalf("categories length = %d, want %d", len(next.categories), len(cats))
	}
}

func TestUpdateHandlesErrMsg(t *testing.T) {
	m := initialModel(nil)

	updated, _ := m.Update(errMsg{err: errors.New("daemon unreachable")})
	next := updated.(model)

	if next.state != stateError {
		t.Fatalf("state = %v, want stateError", next.state)
	}
	if next.err == nil {
		t.Fatalf("expected err to be set")
	}
}

func TestUpdateHandlesPruneResultMsg(t *testing.T) {
	m := initialModel(nil)
	summary := PruneSummary{Results: []PruneResult{{Label: "x", SpaceReclaimed: 42}}}

	updated, _ := m.Update(pruneResultMsg{summary: summary})
	next := updated.(model)

	if next.state != stateSummary {
		t.Fatalf("state = %v, want stateSummary", next.state)
	}
	if next.summary.TotalReclaimed() != 42 {
		t.Fatalf("TotalReclaimed() = %d, want 42", next.summary.TotalReclaimed())
	}
}

func TestAnySelected(t *testing.T) {
	m := initialModel(nil)
	if m.anySelected() {
		t.Fatalf("anySelected() = true, want false on a fresh model")
	}

	m.selected[CategoryBuildCache] = false
	if m.anySelected() {
		t.Fatalf("anySelected() = true, want false when the only entry is false")
	}

	m.selected[CategoryBuildCache] = true
	if !m.anySelected() {
		t.Fatalf("anySelected() = false, want true")
	}
}

func TestTotalAndSelectedSize(t *testing.T) {
	m := newBrowsingModel()

	if got := m.totalSize(); got != 1000 {
		t.Fatalf("totalSize() = %d, want 1000", got)
	}

	m.selected[CategoryDanglingImages] = true
	m.selected[CategoryOrphanVolumes] = true

	if got := m.selectedSize(); got != 400 {
		t.Fatalf("selectedSize() = %d, want 400", got)
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{1, ""},
		{2, "s"},
		{42, "s"},
	}

	for _, tt := range tests {
		if got := plural(tt.n); got != tt.want {
			t.Errorf("plural(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{12 * time.Hour, "moins d'1 jour"},
		{24 * time.Hour, "1 jour"},
		{3 * 24 * time.Hour, "3 jours"},
	}

	for _, tt := range tests {
		if got := formatAge(tt.d); got != tt.want {
			t.Errorf("formatAge(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

// update is a small test helper mirroring what Bubble Tea does internally:
// it runs Update and type-asserts the result back to our concrete model.
func update(m model, msg tea.Msg) (model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(model), cmd
}
