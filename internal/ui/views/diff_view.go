package views

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/a1yama/tig-gh/internal/ui/components"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FetchDiffUseCase defines the interface for fetching diff
type FetchDiffUseCase interface {
	Execute(ctx context.Context, owner, repo string, prNumber int) (string, error)
}

// DiffLineType represents the type of a diff line
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdded
	DiffLineDeleted
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type       DiffLineType
	Content    string
	OldLineNum int
	NewLineNum int
}

// DiffFile represents a file in a diff
type DiffFile struct {
	OldPath string
	NewPath string
	Lines   []DiffLine
}

// diffLoadedMsg is sent when diff is loaded
type diffLoadedMsg struct {
	diff string
	err  error
}

// DiffView is the model for the diff view
type DiffView struct {
	fetchDiffUseCase FetchDiffUseCase
	owner            string
	repo             string
	prNumber         int
	files            []DiffFile
	currentFile      int
	scroll           int
	loading          bool
	err              error
	width            int
	height           int
	statusBar        *components.StatusBar
}

// NewDiffView creates a new diff view
func NewDiffView() *DiffView {
	return &DiffView{
		fetchDiffUseCase: nil,
		owner:            "",
		repo:             "",
		prNumber:         0,
		files:            []DiffFile{},
		currentFile:      0,
		scroll:           0,
		loading:          false,
		statusBar:        components.NewStatusBar(),
	}
}

// NewDiffViewWithUseCase creates a new diff view with UseCase
func NewDiffViewWithUseCase(fetchDiffUseCase FetchDiffUseCase, owner, repo string, prNumber int) *DiffView {
	return &DiffView{
		fetchDiffUseCase: fetchDiffUseCase,
		owner:            owner,
		repo:             repo,
		prNumber:         prNumber,
		files:            []DiffFile{},
		currentFile:      0,
		scroll:           0,
		loading:          true, // Start in loading state
		statusBar:        components.NewStatusBar(),
	}
}

// Init initializes the diff view
func (m *DiffView) Init() tea.Cmd {
	if m.fetchDiffUseCase != nil {
		return m.fetchDiff()
	}
	return nil
}

// Update handles messages
func (m *DiffView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case diffLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.files = []DiffFile{}
		} else {
			m.err = nil
			m.files = parseDiff(msg.diff)
			// Reset cursor if it's out of bounds
			if m.currentFile >= len(m.files) && len(m.files) > 0 {
				m.currentFile = len(m.files) - 1
			} else if len(m.files) == 0 {
				m.currentFile = 0
			}
			m.scroll = 0
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(msg.Width, 1)
		return m, nil
	}

	return m, nil
}

// fetchDiff fetches diff from the API
func (m *DiffView) fetchDiff() tea.Cmd {
	return func() tea.Msg {
		if m.fetchDiffUseCase == nil {
			return diffLoadedMsg{
				diff: "",
				err:  fmt.Errorf("fetch diff use case not initialized"),
			}
		}

		diff, err := m.fetchDiffUseCase.Execute(context.Background(), m.owner, m.repo, m.prNumber)
		return diffLoadedMsg{
			diff: diff,
			err:  err,
		}
	}
}

// handleKeyPress handles keyboard input
func (m *DiffView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "j", "down":
		// Scroll down
		if len(m.files) > 0 && m.currentFile < len(m.files) {
			maxScroll := len(m.files[m.currentFile].Lines) - 1
			if m.scroll < maxScroll {
				m.scroll++
			}
		}
		return m, nil

	case "k", "up":
		// Scroll up
		if m.scroll > 0 {
			m.scroll--
		}
		return m, nil

	case "n":
		// Next file
		if m.currentFile < len(m.files)-1 {
			m.currentFile++
			m.scroll = 0 // Reset scroll when changing files
		}
		return m, nil

	case "p":
		// Previous file
		if m.currentFile > 0 {
			m.currentFile--
			m.scroll = 0 // Reset scroll when changing files
		}
		return m, nil

	case "g":
		// Go to top
		m.scroll = 0
		return m, nil

	case "G":
		// Go to bottom
		if len(m.files) > 0 && m.currentFile < len(m.files) {
			m.scroll = len(m.files[m.currentFile].Lines) - 1
			if m.scroll < 0 {
				m.scroll = 0
			}
		}
		return m, nil
	}

	return m, nil
}

// View renders the diff view
func (m *DiffView) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	// Header
	header := m.renderHeader()
	s.WriteString(header)
	s.WriteString("\n")

	// Diff content or error/loading state
	if m.loading {
		s.WriteString(m.renderLoading())
	} else if m.err != nil {
		s.WriteString(m.renderError())
	} else if len(m.files) == 0 {
		s.WriteString(m.renderEmpty())
	} else {
		s.WriteString(m.renderDiff())
	}

	// Status bar
	s.WriteString("\n")
	m.updateStatusBar()
	s.WriteString(m.statusBar.View())

	return s.String()
}

// renderHeader renders the view header
func (m *DiffView) renderHeader() string {
	title := styles.HeaderStyle.Render(fmt.Sprintf("Diff: PR #%d", m.prNumber))
	if len(m.files) > 0 {
		fileInfo := styles.MutedStyle.Render(fmt.Sprintf("(%d/%d files)", m.currentFile+1, len(m.files)))
		return lipgloss.JoinHorizontal(lipgloss.Top, title, " ", fileInfo)
	}
	return title
}

// renderDiff renders the diff content
func (m *DiffView) renderDiff() string {
	if m.currentFile >= len(m.files) {
		return ""
	}

	file := m.files[m.currentFile]
	var s strings.Builder

	// File header
	fileHeader := m.renderFileHeader()
	s.WriteString(fileHeader)
	s.WriteString("\n")

	// Calculate available height for diff (total - header - file header - status bar - margins)
	availableHeight := m.height - 5

	// Calculate visible range
	startIdx := m.scroll
	endIdx := startIdx + availableHeight
	if endIdx > len(file.Lines) {
		endIdx = len(file.Lines)
	}

	// Render visible lines
	for i := startIdx; i < endIdx; i++ {
		line := m.renderDiffLine(file.Lines[i])
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderFileHeader renders a file header
func (m *DiffView) renderFileHeader() string {
	if m.currentFile >= len(m.files) {
		return ""
	}

	file := m.files[m.currentFile]
	path := file.NewPath
	if file.NewPath != file.OldPath && file.OldPath != "" {
		path = fmt.Sprintf("%s → %s", file.OldPath, file.NewPath)
	}

	return styles.TitleStyle.Render(path)
}

// renderDiffLine renders a single diff line
func (m *DiffView) renderDiffLine(line DiffLine) string {
	// Line number
	lineNum := ""
	if line.Type == DiffLineAdded {
		lineNum = fmt.Sprintf("  +%-4d", line.NewLineNum)
	} else if line.Type == DiffLineDeleted {
		lineNum = fmt.Sprintf("  -%-4d", line.OldLineNum)
	} else {
		lineNum = fmt.Sprintf("   %-4d", line.NewLineNum)
	}

	// Apply style based on line type
	var styledContent string
	switch line.Type {
	case DiffLineAdded:
		styledContent = styles.AddedLineStyle.Render("+" + line.Content)
	case DiffLineDeleted:
		styledContent = styles.DeletedLineStyle.Render("-" + line.Content)
	default:
		styledContent = styles.ContextLineStyle.Render(" " + line.Content)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.MutedStyle.Render(lineNum),
		" ",
		styledContent,
	)
}

// renderLoading renders a loading state
func (m *DiffView) renderLoading() string {
	return styles.LoadingStyle.Render("Loading diff...")
}

// renderError renders an error state
func (m *DiffView) renderError() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
}

// renderEmpty renders an empty state
func (m *DiffView) renderEmpty() string {
	return styles.MutedStyle.Render("No diff available")
}

// updateStatusBar updates the status bar with current state
func (m *DiffView) updateStatusBar() {
	m.statusBar.ClearItems()

	// Set mode
	m.statusBar.SetMode("Diff")

	// Add current position
	if len(m.files) > 0 && m.currentFile < len(m.files) {
		file := m.files[m.currentFile]
		if len(file.Lines) > 0 {
			position := fmt.Sprintf("%d/%d lines", m.scroll+1, len(file.Lines))
			m.statusBar.AddItem("", position)
		}
		filePosition := fmt.Sprintf("file %d/%d", m.currentFile+1, len(m.files))
		m.statusBar.AddItem("", filePosition)
	}

	// Add PR info
	if m.prNumber > 0 {
		m.statusBar.AddItem("PR", fmt.Sprintf("#%d", m.prNumber))
	}

	// Add repository info
	if m.owner != "" && m.repo != "" {
		m.statusBar.AddItem("Repo", fmt.Sprintf("%s/%s", m.owner, m.repo))
	}

	// Add key hints
	m.statusBar.AddItem("", "j/k: scroll | n/p: file | q: quit")
}

// parseDiff parses a unified diff string into DiffFile structures
func parseDiff(diffText string) []DiffFile {
	if diffText == "" {
		return []DiffFile{}
	}

	files := []DiffFile{}
	lines := strings.Split(diffText, "\n")

	var currentFile *DiffFile
	oldLineNum := 0
	newLineNum := 0

	// Regex patterns
	fileHeaderPattern := regexp.MustCompile(`^diff --git a/(.+) b/(.+)$`)
	oldFilePattern := regexp.MustCompile(`^--- a/(.+)$`)
	newFilePattern := regexp.MustCompile(`^\+\+\+ b/(.+)$`)
	hunkHeaderPattern := regexp.MustCompile(`^@@ -(\d+),?\d* \+(\d+),?\d* @@`)

	for _, line := range lines {
		// Check for file header
		if matches := fileHeaderPattern.FindStringSubmatch(line); matches != nil {
			if currentFile != nil {
				files = append(files, *currentFile)
			}
			currentFile = &DiffFile{
				OldPath: matches[1],
				NewPath: matches[2],
				Lines:   []DiffLine{},
			}
			continue
		}

		// Check for old file path
		if matches := oldFilePattern.FindStringSubmatch(line); matches != nil && currentFile != nil {
			currentFile.OldPath = matches[1]
			continue
		}

		// Check for new file path
		if matches := newFilePattern.FindStringSubmatch(line); matches != nil && currentFile != nil {
			currentFile.NewPath = matches[1]
			continue
		}

		// Check for hunk header
		if matches := hunkHeaderPattern.FindStringSubmatch(line); matches != nil {
			if len(matches) >= 3 {
				fmt.Sscanf(matches[1], "%d", &oldLineNum)
				fmt.Sscanf(matches[2], "%d", &newLineNum)
			}
			continue
		}

		// Parse diff lines
		if currentFile != nil && len(line) > 0 {
			firstChar := line[0]
			content := ""
			if len(line) > 1 {
				content = line[1:]
			}

			var diffLine DiffLine
			switch firstChar {
			case '+':
				diffLine = DiffLine{
					Type:       DiffLineAdded,
					Content:    content,
					OldLineNum: 0,
					NewLineNum: newLineNum,
				}
				newLineNum++
			case '-':
				diffLine = DiffLine{
					Type:       DiffLineDeleted,
					Content:    content,
					OldLineNum: oldLineNum,
					NewLineNum: 0,
				}
				oldLineNum++
			case ' ':
				diffLine = DiffLine{
					Type:       DiffLineContext,
					Content:    content,
					OldLineNum: oldLineNum,
					NewLineNum: newLineNum,
				}
				oldLineNum++
				newLineNum++
			default:
				// Skip lines that don't match the expected format
				continue
			}

			currentFile.Lines = append(currentFile.Lines, diffLine)
		}
	}

	// Add the last file
	if currentFile != nil {
		files = append(files, *currentFile)
	}

	return files
}
