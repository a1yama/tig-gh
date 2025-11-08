package components

import (
	"testing"

	"github.com/a1yama/tig-gh/internal/domain/models"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewFilterModal(t *testing.T) {
	fm := NewFilterModal()

	if fm == nil {
		t.Fatal("NewFilterModal should not return nil")
	}

	if fm.visible {
		t.Error("Expected filter modal to be hidden initially")
	}

	if fm.state != models.IssueStateOpen {
		t.Errorf("Expected default state to be open, got %s", fm.state)
	}

	if fm.selectedLabels == nil {
		t.Error("Expected selectedLabels to be initialized")
	}
}

func TestFilterModal_Show(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()

	if !fm.visible {
		t.Error("Expected filter modal to be visible after Show()")
	}
}

func TestFilterModal_Hide(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.Hide()

	if fm.visible {
		t.Error("Expected filter modal to be hidden after Hide()")
	}
}

func TestFilterModal_IsVisible(t *testing.T) {
	fm := NewFilterModal()

	if fm.IsVisible() {
		t.Error("Expected IsVisible() to return false initially")
	}

	fm.Show()

	if !fm.IsVisible() {
		t.Error("Expected IsVisible() to return true after Show()")
	}
}

func TestFilterModal_SetState(t *testing.T) {
	fm := NewFilterModal()

	states := []models.IssueState{
		models.IssueStateOpen,
		models.IssueStateClosed,
		models.IssueStateAll,
	}

	for _, state := range states {
		fm.SetState(state)
		if fm.state != state {
			t.Errorf("Expected state to be %s, got %s", state, fm.state)
		}
	}
}

func TestFilterModal_GetState(t *testing.T) {
	fm := NewFilterModal()
	fm.SetState(models.IssueStateClosed)

	if got := fm.GetState(); got != models.IssueStateClosed {
		t.Errorf("Expected GetState() to return %s, got %s", models.IssueStateClosed, got)
	}
}

func TestFilterModal_SetLabels(t *testing.T) {
	fm := NewFilterModal()

	labels := []string{"bug", "feature", "documentation"}
	fm.SetLabels(labels)

	// Should be able to retrieve them
	selected := fm.GetSelectedLabels()
	if len(selected) != 0 {
		t.Error("Expected no labels selected initially after SetLabels")
	}
}

func TestFilterModal_ToggleLabel(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature"})

	// Toggle bug on
	fm.ToggleLabel("bug")
	selected := fm.GetSelectedLabels()

	if len(selected) != 1 || selected[0] != "bug" {
		t.Errorf("Expected [bug], got %v", selected)
	}

	// Toggle bug off
	fm.ToggleLabel("bug")
	selected = fm.GetSelectedLabels()

	if len(selected) != 0 {
		t.Errorf("Expected empty list, got %v", selected)
	}

	// Toggle multiple labels
	fm.ToggleLabel("bug")
	fm.ToggleLabel("feature")
	selected = fm.GetSelectedLabels()

	if len(selected) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(selected))
	}
}

func TestFilterModal_GetSelectedLabels(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature", "docs"})

	fm.ToggleLabel("bug")
	fm.ToggleLabel("docs")

	selected := fm.GetSelectedLabels()

	expectedCount := 2
	if len(selected) != expectedCount {
		t.Errorf("Expected %d labels, got %d", expectedCount, len(selected))
	}

	hasLabel := func(labels []string, label string) bool {
		for _, l := range labels {
			if l == label {
				return true
			}
		}
		return false
	}

	if !hasLabel(selected, "bug") {
		t.Error("Expected selected labels to include 'bug'")
	}

	if !hasLabel(selected, "docs") {
		t.Error("Expected selected labels to include 'docs'")
	}

	if hasLabel(selected, "feature") {
		t.Error("Expected selected labels not to include 'feature'")
	}
}

func TestFilterModal_SetSort(t *testing.T) {
	fm := NewFilterModal()

	sorts := []models.IssueSort{
		models.IssueSortCreated,
		models.IssueSortUpdated,
		models.IssueSortComments,
	}

	for _, sort := range sorts {
		fm.SetSort(sort)
		if fm.sort != sort {
			t.Errorf("Expected sort to be %s, got %s", sort, fm.sort)
		}
	}
}

func TestFilterModal_GetSort(t *testing.T) {
	fm := NewFilterModal()
	fm.SetSort(models.IssueSortComments)

	if got := fm.GetSort(); got != models.IssueSortComments {
		t.Errorf("Expected GetSort() to return %s, got %s", models.IssueSortComments, got)
	}
}

func TestFilterModal_SetDirection(t *testing.T) {
	fm := NewFilterModal()

	directions := []models.SortDirection{
		models.SortDirectionAsc,
		models.SortDirectionDesc,
	}

	for _, dir := range directions {
		fm.SetDirection(dir)
		if fm.direction != dir {
			t.Errorf("Expected direction to be %s, got %s", dir, fm.direction)
		}
	}
}

func TestFilterModal_GetDirection(t *testing.T) {
	fm := NewFilterModal()
	fm.SetDirection(models.SortDirectionAsc)

	if got := fm.GetDirection(); got != models.SortDirectionAsc {
		t.Errorf("Expected GetDirection() to return %s, got %s", models.SortDirectionAsc, got)
	}
}

func TestFilterModal_Reset(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature"})

	// Change all settings
	fm.SetState(models.IssueStateClosed)
	fm.ToggleLabel("bug")
	fm.SetSort(models.IssueSortComments)
	fm.SetDirection(models.SortDirectionAsc)

	// Reset
	fm.Reset()

	// Check defaults are restored
	if fm.state != models.IssueStateOpen {
		t.Errorf("Expected state to be reset to open, got %s", fm.state)
	}

	if len(fm.GetSelectedLabels()) != 0 {
		t.Error("Expected selected labels to be cleared")
	}

	if fm.sort != models.IssueSortUpdated {
		t.Errorf("Expected sort to be reset to updated, got %s", fm.sort)
	}

	if fm.direction != models.SortDirectionDesc {
		t.Errorf("Expected direction to be reset to desc, got %s", fm.direction)
	}
}

func TestFilterModal_Update_Navigation(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.SetLabels([]string{"bug", "feature", "docs"})

	// Test down navigation
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	fm.Update(downMsg)

	if fm.cursor != 1 {
		t.Errorf("Expected cursor at 1 after down, got %d", fm.cursor)
	}

	// Test up navigation
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	fm.Update(upMsg)

	if fm.cursor != 0 {
		t.Errorf("Expected cursor at 0 after up, got %d", fm.cursor)
	}

	// Test bounds - up at top should stay at top
	fm.Update(upMsg)
	if fm.cursor != 0 {
		t.Error("Expected cursor to stay at 0 when already at top")
	}
}

func TestFilterModal_Update_Selection(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()

	// Simulate selecting a filter option with Enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	fm.Update(enterMsg)

	// Behavior depends on implementation
	// This test ensures Update handles Enter key without panic
}

func TestFilterModal_Update_Escape(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	fm.Update(escMsg)

	if fm.visible {
		t.Error("Expected filter modal to be hidden after Escape")
	}
}

func TestFilterModal_Update_WhenHidden(t *testing.T) {
	fm := NewFilterModal()
	// Modal is hidden

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	fm.Update(downMsg)

	// Should not process keys when hidden
	if fm.cursor != 0 {
		t.Error("Expected no cursor movement when modal is hidden")
	}
}

func TestFilterModal_SetSize(t *testing.T) {
	fm := NewFilterModal()

	fm.SetSize(80, 24)

	if fm.width != 80 {
		t.Errorf("Expected width 80, got %d", fm.width)
	}

	if fm.height != 24 {
		t.Errorf("Expected height 24, got %d", fm.height)
	}
}

func TestFilterModal_View(t *testing.T) {
	fm := NewFilterModal()
	fm.SetSize(80, 24)

	// Test hidden view
	hiddenView := fm.View()
	if hiddenView != "" {
		t.Error("Expected empty view when modal is hidden")
	}

	// Test visible view
	fm.Show()
	fm.SetLabels([]string{"bug", "feature"})
	visibleView := fm.View()

	if visibleView == "" {
		t.Error("Expected non-empty view when modal is visible")
	}
}

func TestFilterModal_GetOptions(t *testing.T) {
	fm := NewFilterModal()
	fm.SetState(models.IssueStateClosed)
	fm.SetLabels([]string{"bug", "feature"})
	fm.ToggleLabel("bug")
	fm.SetSort(models.IssueSortCreated)
	fm.SetDirection(models.SortDirectionAsc)

	opts := fm.GetOptions()

	if opts.State != models.IssueStateClosed {
		t.Errorf("Expected state %s, got %s", models.IssueStateClosed, opts.State)
	}

	if len(opts.Labels) != 1 || opts.Labels[0] != "bug" {
		t.Errorf("Expected labels [bug], got %v", opts.Labels)
	}

	if opts.Sort != models.IssueSortCreated {
		t.Errorf("Expected sort %s, got %s", models.IssueSortCreated, opts.Sort)
	}

	if opts.Direction != models.SortDirectionAsc {
		t.Errorf("Expected direction %s, got %s", models.SortDirectionAsc, opts.Direction)
	}
}

func TestFilterModal_ApplyOptions(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature", "docs"})

	opts := &models.IssueOptions{
		State:     models.IssueStateClosed,
		Labels:    []string{"bug", "docs"},
		Sort:      models.IssueSortComments,
		Direction: models.SortDirectionAsc,
	}

	fm.ApplyOptions(opts)

	if fm.GetState() != models.IssueStateClosed {
		t.Errorf("Expected state %s, got %s", models.IssueStateClosed, fm.GetState())
	}

	selected := fm.GetSelectedLabels()
	if len(selected) != 2 {
		t.Errorf("Expected 2 selected labels, got %d", len(selected))
	}

	if fm.GetSort() != models.IssueSortComments {
		t.Errorf("Expected sort %s, got %s", models.IssueSortComments, fm.GetSort())
	}

	if fm.GetDirection() != models.SortDirectionAsc {
		t.Errorf("Expected direction %s, got %s", models.SortDirectionAsc, fm.GetDirection())
	}
}

func TestFilterModal_ApplyOptions_Nil(t *testing.T) {
	fm := NewFilterModal()
	fm.SetState(models.IssueStateClosed)

	// Should not panic with nil
	fm.ApplyOptions(nil)

	// State should remain unchanged
	if fm.GetState() != models.IssueStateClosed {
		t.Error("State should not change when applying nil options")
	}
}

func TestFilterModal_HandleSelection_State(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()

	// Select "Open" (cursor 0)
	fm.cursor = 0
	fm.handleSelection()
	if fm.state != models.IssueStateOpen {
		t.Errorf("Expected state Open, got %s", fm.state)
	}

	// Select "Closed" (cursor 1)
	fm.cursor = 1
	fm.handleSelection()
	if fm.state != models.IssueStateClosed {
		t.Errorf("Expected state Closed, got %s", fm.state)
	}

	// Select "All" (cursor 2)
	fm.cursor = 2
	fm.handleSelection()
	if fm.state != models.IssueStateAll {
		t.Errorf("Expected state All, got %s", fm.state)
	}
}

func TestFilterModal_HandleSelection_Labels(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.SetLabels([]string{"bug", "feature"})

	// Select first label (cursor 3, after 3 state options)
	fm.cursor = 3
	fm.handleSelection()

	selected := fm.GetSelectedLabels()
	if len(selected) != 1 || selected[0] != "bug" {
		t.Errorf("Expected [bug], got %v", selected)
	}

	// Toggle it off
	fm.handleSelection()
	selected = fm.GetSelectedLabels()
	if len(selected) != 0 {
		t.Errorf("Expected empty, got %v", selected)
	}
}

func TestFilterModal_HandleSelection_Sort(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.SetLabels([]string{"bug"}) // 1 label

	// Sort options start at cursor 3 (states) + 1 (label) = 4

	// Select "Created" (cursor 4)
	fm.cursor = 4
	fm.handleSelection()
	if fm.sort != models.IssueSortCreated {
		t.Errorf("Expected sort Created, got %s", fm.sort)
	}

	// Select "Updated" (cursor 5)
	fm.cursor = 5
	fm.handleSelection()
	if fm.sort != models.IssueSortUpdated {
		t.Errorf("Expected sort Updated, got %s", fm.sort)
	}

	// Select "Comments" (cursor 6)
	fm.cursor = 6
	fm.handleSelection()
	if fm.sort != models.IssueSortComments {
		t.Errorf("Expected sort Comments, got %s", fm.sort)
	}
}

func TestFilterModal_HandleSelection_Direction(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.SetLabels([]string{"bug"}) // 1 label

	// Direction options start at cursor 3 (states) + 1 (label) + 3 (sorts) = 7

	// Select "Ascending" (cursor 7)
	fm.cursor = 7
	fm.handleSelection()
	if fm.direction != models.SortDirectionAsc {
		t.Errorf("Expected direction Asc, got %s", fm.direction)
	}

	// Select "Descending" (cursor 8)
	fm.cursor = 8
	fm.handleSelection()
	if fm.direction != models.SortDirectionDesc {
		t.Errorf("Expected direction Desc, got %s", fm.direction)
	}
}

func TestFilterModal_GetMaxCursor(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature", "docs"})

	// 3 states + 3 labels + 3 sorts + 2 directions - 1 = 10
	maxCursor := fm.getMaxCursor()
	expected := 3 + 3 + 3 + 2 - 1
	if maxCursor != expected {
		t.Errorf("Expected max cursor %d, got %d", expected, maxCursor)
	}
}

func TestFilterModal_Update_DownBeyondMax(t *testing.T) {
	fm := NewFilterModal()
	fm.Show()
	fm.SetLabels([]string{"bug"})

	maxCursor := fm.getMaxCursor()
	fm.cursor = maxCursor

	// Try to go down beyond max
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	fm.Update(downMsg)

	// Should stay at max
	if fm.cursor != maxCursor {
		t.Errorf("Cursor should stay at max %d, got %d", maxCursor, fm.cursor)
	}
}

func TestFilterModal_SetLabels_UpdatesExistingSelections(t *testing.T) {
	fm := NewFilterModal()
	fm.SetLabels([]string{"bug", "feature", "docs"})
	fm.ToggleLabel("bug")
	fm.ToggleLabel("feature")

	// Now set new labels that only include "bug"
	fm.SetLabels([]string{"bug", "critical"})

	selected := fm.GetSelectedLabels()
	// Should only have "bug" now since "feature" is no longer available
	if len(selected) != 1 || selected[0] != "bug" {
		t.Errorf("Expected [bug], got %v", selected)
	}
}

func TestFilterModal_View_WithLabels(t *testing.T) {
	fm := NewFilterModal()
	fm.SetSize(80, 24)
	fm.SetLabels([]string{"bug", "feature"})
	fm.Show()

	view := fm.View()
	if view == "" {
		t.Error("View should not be empty when visible with labels")
	}
}

func TestFilterModal_RenderSections(t *testing.T) {
	fm := NewFilterModal()
	fm.SetSize(80, 24)
	fm.SetLabels([]string{"bug"})
	fm.Show()

	// Test state section rendering
	currentIndex := 0
	stateView := fm.renderStateSection(&currentIndex)
	if stateView == "" {
		t.Error("State section should not be empty")
	}
	if currentIndex != 3 {
		t.Errorf("Expected currentIndex to be 3 after states, got %d", currentIndex)
	}

	// Test labels section rendering
	labelsView := fm.renderLabelsSection(&currentIndex)
	if labelsView == "" {
		t.Error("Labels section should not be empty")
	}
	if currentIndex != 4 {
		t.Errorf("Expected currentIndex to be 4 after labels, got %d", currentIndex)
	}

	// Test sort section rendering
	sortView := fm.renderSortSection(&currentIndex)
	if sortView == "" {
		t.Error("Sort section should not be empty")
	}
	if currentIndex != 7 {
		t.Errorf("Expected currentIndex to be 7 after sorts, got %d", currentIndex)
	}

	// Test direction section rendering
	directionView := fm.renderDirectionSection(&currentIndex)
	if directionView == "" {
		t.Error("Direction section should not be empty")
	}
	if currentIndex != 9 {
		t.Errorf("Expected currentIndex to be 9 after directions, got %d", currentIndex)
	}
}
