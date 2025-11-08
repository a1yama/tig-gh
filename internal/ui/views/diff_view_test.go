package views

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/a1yama/tig-gh/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// mockFetchDiffUseCase is a mock implementation for testing
type mockFetchDiffUseCase struct {
	executeFunc func(ctx context.Context, owner, repo string, prNumber int) (string, error)
}

func (m *mockFetchDiffUseCase) Execute(ctx context.Context, owner, repo string, prNumber int) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, owner, repo, prNumber)
	}
	return "", nil
}

// サンプル差分データ
const sampleDiff = `diff --git a/src/components/Dashboard.tsx b/src/components/Dashboard.tsx
index 1234567..abcdefg 100644
--- a/src/components/Dashboard.tsx
+++ b/src/components/Dashboard.tsx
@@ -1,10 +1,11 @@
 import React from 'react';
+import { Chart } from './Chart';

 export const Dashboard = () => {
   return (
     <div className="dashboard">
+      <Chart data={data} />
-      <OldComponent />
     </div>
   );
 };
`

func TestDiffView_Init(t *testing.T) {
	tests := []struct {
		name          string
		mockUseCase   *mockFetchDiffUseCase
		checkCmd      func(t *testing.T, cmd tea.Cmd)
		expectLoading bool
	}{
		{
			name: "init triggers diff fetch",
			mockUseCase: &mockFetchDiffUseCase{
				executeFunc: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
					return sampleDiff, nil
				},
			},
			checkCmd: func(t *testing.T, cmd tea.Cmd) {
				if cmd == nil {
					t.Error("expected non-nil command from Init()")
				}
			},
			expectLoading: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := NewDiffViewWithUseCase(tt.mockUseCase, "testowner", "testrepo", 123)
			cmd := view.Init()

			if tt.checkCmd != nil {
				tt.checkCmd(t, cmd)
			}

			if view.loading != tt.expectLoading {
				t.Errorf("expected loading=%v, got %v", tt.expectLoading, view.loading)
			}
		})
	}
}

func TestDiffView_Update_DiffLoaded(t *testing.T) {
	tests := []struct {
		name          string
		msg           tea.Msg
		initialState  *DiffView
		checkState    func(t *testing.T, view *DiffView)
		expectLoading bool
	}{
		{
			name: "successful diff load",
			msg: diffLoadedMsg{
				diff: sampleDiff,
				err:  nil,
			},
			initialState: &DiffView{
				loading:  true,
				owner:    "testowner",
				repo:     "testrepo",
				prNumber: 123,
				scroll:   0,
			},
			checkState: func(t *testing.T, view *DiffView) {
				if view.loading {
					t.Error("expected loading to be false after successful load")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.files) == 0 {
					t.Error("expected files to be parsed from diff")
				}
			},
			expectLoading: false,
		},
		{
			name: "error during load",
			msg: diffLoadedMsg{
				diff: "",
				err:  errors.New("API error"),
			},
			initialState: &DiffView{
				loading:  true,
				owner:    "testowner",
				repo:     "testrepo",
				prNumber: 123,
			},
			checkState: func(t *testing.T, view *DiffView) {
				if view.loading {
					t.Error("expected loading to be false after error")
				}
				if view.err == nil {
					t.Error("expected error to be set")
				}
				if view.err.Error() != "API error" {
					t.Errorf("expected error 'API error', got '%v'", view.err)
				}
			},
			expectLoading: false,
		},
		{
			name: "empty diff",
			msg: diffLoadedMsg{
				diff: "",
				err:  nil,
			},
			initialState: &DiffView{
				loading:  true,
				owner:    "testowner",
				repo:     "testrepo",
				prNumber: 123,
			},
			checkState: func(t *testing.T, view *DiffView) {
				if view.loading {
					t.Error("expected loading to be false")
				}
				if view.err != nil {
					t.Errorf("expected no error, got %v", view.err)
				}
				if len(view.files) != 0 {
					t.Errorf("expected 0 files for empty diff, got %d", len(view.files))
				}
			},
			expectLoading: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.initialState
			view.statusBar = components.NewStatusBar()
			_, _ = view.Update(tt.msg)

			if tt.checkState != nil {
				tt.checkState(t, view)
			}

			if view.loading != tt.expectLoading {
				t.Errorf("expected loading=%v, got %v", tt.expectLoading, view.loading)
			}
		})
	}
}

func TestDiffView_Update_ScrollKeys(t *testing.T) {
	tests := []struct {
		name           string
		initialScroll  int
		key            string
		expectedScroll int
		totalLines     int
	}{
		{
			name:           "scroll down with j",
			initialScroll:  0,
			key:            "j",
			expectedScroll: 1,
			totalLines:     20,
		},
		{
			name:           "scroll down with down arrow",
			initialScroll:  0,
			key:            "down",
			expectedScroll: 1,
			totalLines:     20,
		},
		{
			name:           "scroll up with k",
			initialScroll:  5,
			key:            "k",
			expectedScroll: 4,
			totalLines:     20,
		},
		{
			name:           "scroll up with up arrow",
			initialScroll:  5,
			key:            "up",
			expectedScroll: 4,
			totalLines:     20,
		},
		{
			name:           "cannot scroll up past 0",
			initialScroll:  0,
			key:            "k",
			expectedScroll: 0,
			totalLines:     20,
		},
		{
			name:           "cannot scroll down past max",
			initialScroll:  19,
			key:            "j",
			expectedScroll: 19,
			totalLines:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := &DiffView{
				loading:    false,
				scroll:     tt.initialScroll,
				files:      []DiffFile{{Lines: make([]DiffLine, tt.totalLines)}},
				currentFile: 0,
				statusBar:  components.NewStatusBar(),
			}

			var msg tea.Msg
			if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			view.Update(msg)

			if view.scroll != tt.expectedScroll {
				t.Errorf("expected scroll at %d, got %d", tt.expectedScroll, view.scroll)
			}
		})
	}
}

func TestDiffView_Update_FileNavigation(t *testing.T) {
	tests := []struct {
		name            string
		initialFile     int
		key             string
		expectedFile    int
		totalFiles      int
	}{
		{
			name:         "next file with n",
			initialFile:  0,
			key:          "n",
			expectedFile: 1,
			totalFiles:   3,
		},
		{
			name:         "previous file with p",
			initialFile:  2,
			key:          "p",
			expectedFile: 1,
			totalFiles:   3,
		},
		{
			name:         "cannot go to next file past last",
			initialFile:  2,
			key:          "n",
			expectedFile: 2,
			totalFiles:   3,
		},
		{
			name:         "cannot go to previous file past first",
			initialFile:  0,
			key:          "p",
			expectedFile: 0,
			totalFiles:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make([]DiffFile, tt.totalFiles)
			for i := range files {
				files[i] = DiffFile{
					OldPath: "file" + string(rune('1'+i)),
					NewPath: "file" + string(rune('1'+i)),
					Lines:   []DiffLine{},
				}
			}

			view := &DiffView{
				loading:     false,
				files:       files,
				currentFile: tt.initialFile,
				scroll:      0,
				statusBar:   components.NewStatusBar(),
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			view.Update(msg)

			if view.currentFile != tt.expectedFile {
				t.Errorf("expected currentFile at %d, got %d", tt.expectedFile, view.currentFile)
			}
		})
	}
}

func TestDiffView_Update_FileNavigationResetScroll(t *testing.T) {
	files := []DiffFile{
		{OldPath: "file1.go", NewPath: "file1.go", Lines: []DiffLine{}},
		{OldPath: "file2.go", NewPath: "file2.go", Lines: []DiffLine{}},
	}

	view := &DiffView{
		loading:     false,
		files:       files,
		currentFile: 0,
		scroll:      10, // Start with non-zero scroll
		statusBar:   components.NewStatusBar(),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	view.Update(msg)

	if view.scroll != 0 {
		t.Errorf("expected scroll to be reset to 0 after file navigation, got %d", view.scroll)
	}
}

func TestDiffView_View_States(t *testing.T) {
	tests := []struct {
		name        string
		view        *DiffView
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "loading state",
			view: &DiffView{
				loading:   true,
				width:     80,
				height:    24,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Loading") {
					t.Error("expected loading message in output")
				}
			},
		},
		{
			name: "error state",
			view: &DiffView{
				loading:   false,
				err:       errors.New("API error"),
				width:     80,
				height:    24,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Error") {
					t.Error("expected error message in output")
				}
				if !strings.Contains(output, "API error") {
					t.Error("expected specific error message in output")
				}
			},
		},
		{
			name: "empty diff state",
			view: &DiffView{
				loading:   false,
				files:     []DiffFile{},
				width:     80,
				height:    24,
				statusBar: components.NewStatusBar(),
			},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "No diff") && !strings.Contains(output, "Empty") {
					t.Error("expected empty diff message in output")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.view.View()

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestDiffView_View_DiffColoring(t *testing.T) {
	view := &DiffView{
		loading: false,
		files: []DiffFile{
			{
				OldPath: "test.go",
				NewPath: "test.go",
				Lines: []DiffLine{
					{Type: DiffLineContext, Content: "import React from 'react';", OldLineNum: 1, NewLineNum: 1},
					{Type: DiffLineAdded, Content: "import { Chart } from './Chart';", OldLineNum: 0, NewLineNum: 2},
					{Type: DiffLineDeleted, Content: "import { OldComponent } from './Old';", OldLineNum: 2, NewLineNum: 0},
				},
			},
		},
		currentFile: 0,
		scroll:      0,
		width:       80,
		height:      24,
		statusBar:   components.NewStatusBar(),
	}

	output := view.View()

	// Verify that diff rendering includes the content
	if !strings.Contains(output, "React") {
		t.Error("expected context line content in output")
	}
	if !strings.Contains(output, "Chart") {
		t.Error("expected added line content in output")
	}
	if !strings.Contains(output, "OldComponent") {
		t.Error("expected deleted line content in output")
	}
}

func TestDiffView_ParseDiff(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFiles int
		checkFiles    func(t *testing.T, files []DiffFile)
	}{
		{
			name:          "parse sample diff",
			input:         sampleDiff,
			expectedFiles: 1,
			checkFiles: func(t *testing.T, files []DiffFile) {
				if len(files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(files))
				}
				file := files[0]
				if !strings.Contains(file.NewPath, "Dashboard.tsx") {
					t.Errorf("expected file path to contain 'Dashboard.tsx', got '%s'", file.NewPath)
				}
				if len(file.Lines) == 0 {
					t.Error("expected file to have lines")
				}

				// Check for added and deleted lines
				hasAdded := false
				hasDeleted := false
				for _, line := range file.Lines {
					if line.Type == DiffLineAdded {
						hasAdded = true
					}
					if line.Type == DiffLineDeleted {
						hasDeleted = true
					}
				}
				if !hasAdded {
					t.Error("expected at least one added line")
				}
				if !hasDeleted {
					t.Error("expected at least one deleted line")
				}
			},
		},
		{
			name:          "empty diff",
			input:         "",
			expectedFiles: 0,
			checkFiles: func(t *testing.T, files []DiffFile) {
				if len(files) != 0 {
					t.Errorf("expected 0 files for empty diff, got %d", len(files))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := parseDiff(tt.input)

			if len(files) != tt.expectedFiles {
				t.Errorf("expected %d files, got %d", tt.expectedFiles, len(files))
			}

			if tt.checkFiles != nil {
				tt.checkFiles(t, files)
			}
		})
	}
}

func TestDiffView_WindowSizeMsg(t *testing.T) {
	view := &DiffView{
		loading:   false,
		statusBar: components.NewStatusBar(),
	}

	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	view.Update(msg)

	if view.width != 100 {
		t.Errorf("expected width=100, got %d", view.width)
	}
	if view.height != 30 {
		t.Errorf("expected height=30, got %d", view.height)
	}
}

func TestDiffView_NewDiffView(t *testing.T) {
	view := NewDiffView()

	if view == nil {
		t.Fatal("expected non-nil view")
	}
	if view.loading {
		t.Error("expected loading to be false for new view")
	}
	if view.statusBar == nil {
		t.Error("expected statusBar to be initialized")
	}
}

func TestDiffView_QuitKey(t *testing.T) {
	view := &DiffView{
		loading:   false,
		statusBar: components.NewStatusBar(),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := view.Update(msg)

	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestDiffView_ScrollToTopBottom(t *testing.T) {
	files := []DiffFile{
		{
			OldPath: "test.go",
			NewPath: "test.go",
			Lines: []DiffLine{
				{Type: DiffLineContext, Content: "line 1"},
				{Type: DiffLineContext, Content: "line 2"},
				{Type: DiffLineContext, Content: "line 3"},
				{Type: DiffLineContext, Content: "line 4"},
				{Type: DiffLineContext, Content: "line 5"},
			},
		},
	}

	view := &DiffView{
		loading:     false,
		files:       files,
		currentFile: 0,
		scroll:      2,
		statusBar:   components.NewStatusBar(),
	}

	// Test 'g' key (go to top)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	view.Update(msg)

	if view.scroll != 0 {
		t.Errorf("expected scroll to be 0 after 'g' key, got %d", view.scroll)
	}

	// Test 'G' key (go to bottom)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	view.Update(msg)

	expectedScroll := len(files[0].Lines) - 1
	if view.scroll != expectedScroll {
		t.Errorf("expected scroll to be %d after 'G' key, got %d", expectedScroll, view.scroll)
	}
}

func TestDiffView_InitWithoutUseCase(t *testing.T) {
	view := NewDiffView()
	cmd := view.Init()

	if cmd != nil {
		t.Error("expected nil command when usecase is not set")
	}
}

func TestDiffView_FetchDiffError(t *testing.T) {
	mockUseCase := &mockFetchDiffUseCase{
		executeFunc: func(ctx context.Context, owner, repo string, prNumber int) (string, error) {
			return "", errors.New("network error")
		},
	}

	view := NewDiffViewWithUseCase(mockUseCase, "testowner", "testrepo", 123)
	cmd := view.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	// Execute the command to get the message
	msg := cmd()

	// Update the view with the message
	view.Update(msg)

	if view.err == nil {
		t.Error("expected error to be set")
	}
	if !view.loading && view.err.Error() != "network error" {
		t.Errorf("expected error 'network error', got %v", view.err)
	}
}

func TestDiffView_View_WithMultipleFiles(t *testing.T) {
	view := &DiffView{
		loading: false,
		files: []DiffFile{
			{
				OldPath: "file1.go",
				NewPath: "file1.go",
				Lines: []DiffLine{
					{Type: DiffLineContext, Content: "package main", OldLineNum: 1, NewLineNum: 1},
				},
			},
			{
				OldPath: "file2.go",
				NewPath: "file2.go",
				Lines: []DiffLine{
					{Type: DiffLineAdded, Content: "import fmt", OldLineNum: 0, NewLineNum: 1},
				},
			},
		},
		currentFile: 0,
		scroll:      0,
		width:       80,
		height:      24,
		prNumber:    123,
		statusBar:   components.NewStatusBar(),
	}

	output := view.View()

	if !strings.Contains(output, "file1.go") {
		t.Error("expected file1.go in output")
	}
	if !strings.Contains(output, "1/2 files") {
		t.Error("expected file count in output")
	}
}

func TestDiffView_RenderFileHeader_Renamed(t *testing.T) {
	view := &DiffView{
		loading: false,
		files: []DiffFile{
			{
				OldPath: "old_name.go",
				NewPath: "new_name.go",
				Lines:   []DiffLine{},
			},
		},
		currentFile: 0,
		statusBar:   components.NewStatusBar(),
	}

	header := view.renderFileHeader()

	if !strings.Contains(header, "old_name.go") || !strings.Contains(header, "new_name.go") {
		t.Error("expected both old and new file names in header for renamed file")
	}
}
