package components

import (
	"fmt"
	"strings"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilterOption represents a filter option in the modal
type FilterOption struct {
	Label    string
	Selected bool
}

// FilterModal represents a filter configuration modal
type FilterModal struct {
	visible        bool
	width          int
	height         int
	cursor         int
	state          models.IssueState
	availableLabels []string
	selectedLabels map[string]bool
	sort           models.IssueSort
	direction      models.SortDirection
}

// NewFilterModal creates a new filter modal
func NewFilterModal() *FilterModal {
	return &FilterModal{
		visible:        false,
		cursor:         0,
		state:          models.IssueStateOpen,
		availableLabels: []string{},
		selectedLabels: make(map[string]bool),
		sort:           models.IssueSortUpdated,
		direction:      models.SortDirectionDesc,
	}
}

// Show displays the filter modal
func (f *FilterModal) Show() {
	f.visible = true
}

// Hide hides the filter modal
func (f *FilterModal) Hide() {
	f.visible = false
}

// IsVisible returns true if the modal is visible
func (f *FilterModal) IsVisible() bool {
	return f.visible
}

// SetSize sets the size of the modal
func (f *FilterModal) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// SetState sets the issue state filter
func (f *FilterModal) SetState(state models.IssueState) {
	f.state = state
}

// GetState returns the current state filter
func (f *FilterModal) GetState() models.IssueState {
	return f.state
}

// SetLabels sets the available labels
func (f *FilterModal) SetLabels(labels []string) {
	f.availableLabels = labels
	// Keep existing selections that are still valid
	newSelected := make(map[string]bool)
	for label := range f.selectedLabels {
		for _, availableLabel := range labels {
			if label == availableLabel {
				newSelected[label] = true
				break
			}
		}
	}
	f.selectedLabels = newSelected
}

// ToggleLabel toggles the selection of a label
func (f *FilterModal) ToggleLabel(label string) {
	if f.selectedLabels[label] {
		delete(f.selectedLabels, label)
	} else {
		f.selectedLabels[label] = true
	}
}

// GetSelectedLabels returns the selected labels
func (f *FilterModal) GetSelectedLabels() []string {
	var labels []string
	for label := range f.selectedLabels {
		labels = append(labels, label)
	}
	return labels
}

// SetSort sets the sort field
func (f *FilterModal) SetSort(sort models.IssueSort) {
	f.sort = sort
}

// GetSort returns the current sort field
func (f *FilterModal) GetSort() models.IssueSort {
	return f.sort
}

// SetDirection sets the sort direction
func (f *FilterModal) SetDirection(direction models.SortDirection) {
	f.direction = direction
}

// GetDirection returns the current sort direction
func (f *FilterModal) GetDirection() models.SortDirection {
	return f.direction
}

// Reset resets all filters to default values
func (f *FilterModal) Reset() {
	f.state = models.IssueStateOpen
	f.selectedLabels = make(map[string]bool)
	f.sort = models.IssueSortUpdated
	f.direction = models.SortDirectionDesc
	f.cursor = 0
}

// GetOptions returns the current filter options as IssueOptions
func (f *FilterModal) GetOptions() *models.IssueOptions {
	return &models.IssueOptions{
		State:     f.state,
		Labels:    f.GetSelectedLabels(),
		Sort:      f.sort,
		Direction: f.direction,
	}
}

// ApplyOptions applies the given options to the filter
func (f *FilterModal) ApplyOptions(opts *models.IssueOptions) {
	if opts == nil {
		return
	}

	f.state = opts.State
	f.sort = opts.Sort
	f.direction = opts.Direction

	// Clear and set labels
	f.selectedLabels = make(map[string]bool)
	for _, label := range opts.Labels {
		f.selectedLabels[label] = true
	}
}

// Update handles input events
func (f *FilterModal) Update(msg tea.Msg) {
	if !f.visible {
		return
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if f.cursor > 0 {
				f.cursor--
			}

		case tea.KeyDown:
			// Calculate max cursor position based on visible options
			maxCursor := f.getMaxCursor()
			if f.cursor < maxCursor {
				f.cursor++
			}

		case tea.KeyEnter:
			f.handleSelection()

		case tea.KeyEsc:
			f.Hide()
		}
	}
}

// getMaxCursor returns the maximum cursor position
func (f *FilterModal) getMaxCursor() int {
	// State options (3: open, closed, all)
	stateOptions := 3
	// Label options
	labelOptions := len(f.availableLabels)
	// Sort options (3: created, updated, comments)
	sortOptions := 3
	// Direction options (2: asc, desc)
	directionOptions := 2

	return stateOptions + labelOptions + sortOptions + directionOptions - 1
}

// handleSelection handles the selection at the current cursor position
func (f *FilterModal) handleSelection() {
	position := f.cursor

	// State section (0-2)
	if position >= 0 && position <= 2 {
		switch position {
		case 0:
			f.state = models.IssueStateOpen
		case 1:
			f.state = models.IssueStateClosed
		case 2:
			f.state = models.IssueStateAll
		}
		return
	}

	position -= 3

	// Label section
	if position >= 0 && position < len(f.availableLabels) {
		label := f.availableLabels[position]
		f.ToggleLabel(label)
		return
	}

	position -= len(f.availableLabels)

	// Sort section (0-2)
	if position >= 0 && position <= 2 {
		switch position {
		case 0:
			f.sort = models.IssueSortCreated
		case 1:
			f.sort = models.IssueSortUpdated
		case 2:
			f.sort = models.IssueSortComments
		}
		return
	}

	position -= 3

	// Direction section (0-1)
	if position >= 0 && position <= 1 {
		switch position {
		case 0:
			f.direction = models.SortDirectionAsc
		case 1:
			f.direction = models.SortDirectionDesc
		}
		return
	}
}

// View renders the filter modal
func (f *FilterModal) View() string {
	if !f.visible {
		return ""
	}

	var sections []string
	currentIndex := 0

	// State section
	sections = append(sections, f.renderStateSection(&currentIndex))

	// Labels section
	if len(f.availableLabels) > 0 {
		sections = append(sections, f.renderLabelsSection(&currentIndex))
	}

	// Sort section
	sections = append(sections, f.renderSortSection(&currentIndex))

	// Direction section
	sections = append(sections, f.renderDirectionSection(&currentIndex))

	// Action buttons
	sections = append(sections, f.renderActions())

	content := strings.Join(sections, "\n\n")

	// Wrap in a modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(f.width - 20).
		MaxWidth(60)

	title := styles.HeaderStyle.Render("Filters")

	return lipgloss.Place(
		f.width,
		f.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(title+"\n\n"+content),
	)
}

// renderStateSection renders the state filter section
func (f *FilterModal) renderStateSection(currentIndex *int) string {
	var lines []string
	lines = append(lines, styles.BoldStyle.Render("State:"))

	states := []struct {
		state models.IssueState
		label string
	}{
		{models.IssueStateOpen, "Open"},
		{models.IssueStateClosed, "Closed"},
		{models.IssueStateAll, "All"},
	}

	for _, s := range states {
		cursor := "  "
		if *currentIndex == f.cursor {
			cursor = styles.CursorStyle.Render("▶ ")
		}

		checkbox := "[ ]"
		if f.state == s.state {
			checkbox = "[✓]"
		}

		line := cursor + checkbox + " " + s.label
		if *currentIndex == f.cursor {
			line = styles.SelectedStyle.Render(line)
		}

		lines = append(lines, line)
		*currentIndex++
	}

	return strings.Join(lines, "\n")
}

// renderLabelsSection renders the labels filter section
func (f *FilterModal) renderLabelsSection(currentIndex *int) string {
	var lines []string
	lines = append(lines, styles.BoldStyle.Render("Labels:"))

	for _, label := range f.availableLabels {
		cursor := "  "
		if *currentIndex == f.cursor {
			cursor = styles.CursorStyle.Render("▶ ")
		}

		checkbox := "[ ]"
		if f.selectedLabels[label] {
			checkbox = "[✓]"
		}

		line := cursor + checkbox + " " + label
		if *currentIndex == f.cursor {
			line = styles.SelectedStyle.Render(line)
		}

		lines = append(lines, line)
		*currentIndex++
	}

	return strings.Join(lines, "\n")
}

// renderSortSection renders the sort filter section
func (f *FilterModal) renderSortSection(currentIndex *int) string {
	var lines []string
	lines = append(lines, styles.BoldStyle.Render("Sort by:"))

	sorts := []struct {
		sort  models.IssueSort
		label string
	}{
		{models.IssueSortCreated, "Created"},
		{models.IssueSortUpdated, "Updated"},
		{models.IssueSortComments, "Comments"},
	}

	for _, s := range sorts {
		cursor := "  "
		if *currentIndex == f.cursor {
			cursor = styles.CursorStyle.Render("▶ ")
		}

		checkbox := "( )"
		if f.sort == s.sort {
			checkbox = "(●)"
		}

		line := cursor + checkbox + " " + s.label
		if *currentIndex == f.cursor {
			line = styles.SelectedStyle.Render(line)
		}

		lines = append(lines, line)
		*currentIndex++
	}

	return strings.Join(lines, "\n")
}

// renderDirectionSection renders the direction filter section
func (f *FilterModal) renderDirectionSection(currentIndex *int) string {
	var lines []string
	lines = append(lines, styles.BoldStyle.Render("Direction:"))

	directions := []struct {
		direction models.SortDirection
		label     string
	}{
		{models.SortDirectionAsc, "Ascending"},
		{models.SortDirectionDesc, "Descending"},
	}

	for _, d := range directions {
		cursor := "  "
		if *currentIndex == f.cursor {
			cursor = styles.CursorStyle.Render("▶ ")
		}

		checkbox := "( )"
		if f.direction == d.direction {
			checkbox = "(●)"
		}

		line := cursor + checkbox + " " + d.label
		if *currentIndex == f.cursor {
			line = styles.SelectedStyle.Render(line)
		}

		lines = append(lines, line)
		*currentIndex++
	}

	return strings.Join(lines, "\n")
}

// renderActions renders the action buttons
func (f *FilterModal) renderActions() string {
	help := styles.HelpStyle.Render(
		fmt.Sprintf("%s %s  %s %s  %s %s",
			styles.HelpKeyStyle.Render("↑/↓"),
			"navigate",
			styles.HelpKeyStyle.Render("Enter"),
			"select",
			styles.HelpKeyStyle.Render("Esc"),
			"close",
		),
	)

	return "\n" + help
}
