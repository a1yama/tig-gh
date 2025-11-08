package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSearchInput(t *testing.T) {
	si := NewSearchInput()

	if si == nil {
		t.Fatal("NewSearchInput should not return nil")
	}

	if si.value != "" {
		t.Errorf("Expected initial value to be empty, got %q", si.value)
	}

	if si.active {
		t.Error("Expected search input to be inactive initially")
	}

	if si.cursor != 0 {
		t.Errorf("Expected initial cursor position to be 0, got %d", si.cursor)
	}
}

func TestSearchInput_Activate(t *testing.T) {
	si := NewSearchInput()
	si.Activate()

	if !si.active {
		t.Error("Expected search input to be active after Activate()")
	}
}

func TestSearchInput_Deactivate(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.Deactivate()

	if si.active {
		t.Error("Expected search input to be inactive after Deactivate()")
	}
}

func TestSearchInput_SetValue(t *testing.T) {
	si := NewSearchInput()

	testCases := []struct {
		name  string
		value string
	}{
		{"empty string", ""},
		{"simple text", "bug"},
		{"with space", "bug fix"},
		{"with colon", "author:user"},
		{"complex query", `author:user label:bug state:open "exact match"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			si.SetValue(tc.value)
			if si.value != tc.value {
				t.Errorf("Expected value %q, got %q", tc.value, si.value)
			}
		})
	}
}

func TestSearchInput_GetValue(t *testing.T) {
	si := NewSearchInput()
	testValue := "test search"

	si.SetValue(testValue)
	if got := si.GetValue(); got != testValue {
		t.Errorf("Expected GetValue() to return %q, got %q", testValue, got)
	}
}

func TestSearchInput_Clear(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test")
	si.Clear()

	if si.value != "" {
		t.Errorf("Expected value to be empty after Clear(), got %q", si.value)
	}

	if si.cursor != 0 {
		t.Errorf("Expected cursor to be 0 after Clear(), got %d", si.cursor)
	}
}

func TestSearchInput_Update_CharacterInput(t *testing.T) {
	si := NewSearchInput()
	si.Activate()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"single character", "a", "a"},
		{"number", "1", "1"},
		{"special char", ":", ":"},
		{"space", " ", " "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			si.Clear()
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.input)}
			si.Update(msg)

			if si.value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, si.value)
			}
		})
	}
}

func TestSearchInput_Update_MultipleCharacters(t *testing.T) {
	si := NewSearchInput()
	si.Activate()

	chars := []string{"a", "u", "t", "h", "o", "r", ":", "u", "s", "e", "r"}
	expected := "author:user"

	for _, char := range chars {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)}
		si.Update(msg)
	}

	if si.value != expected {
		t.Errorf("Expected value %q, got %q", expected, si.value)
	}
}

func TestSearchInput_Update_Backspace(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.SetValue("test")
	si.cursor = len(si.value) // Move cursor to end

	// Backspace at end
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	si.Update(msg)

	if si.value != "tes" {
		t.Errorf("Expected value %q after backspace, got %q", "tes", si.value)
	}

	// Multiple backspaces
	si.Update(msg)
	si.Update(msg)

	if si.value != "t" {
		t.Errorf("Expected value %q after multiple backspaces, got %q", "t", si.value)
	}

	// Backspace on empty should not panic
	si.Clear()
	si.Update(msg)

	if si.value != "" {
		t.Errorf("Expected value to remain empty, got %q", si.value)
	}
}

func TestSearchInput_Update_CursorMovement(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.SetValue("test")
	si.cursor = len(si.value)

	// Move left
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	si.Update(leftMsg)
	if si.cursor != 3 {
		t.Errorf("Expected cursor at 3 after left, got %d", si.cursor)
	}

	// Move right
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	si.Update(rightMsg)
	if si.cursor != 4 {
		t.Errorf("Expected cursor at 4 after right, got %d", si.cursor)
	}

	// Move to beginning
	homeMsg := tea.KeyMsg{Type: tea.KeyHome}
	si.Update(homeMsg)
	if si.cursor != 0 {
		t.Errorf("Expected cursor at 0 after home, got %d", si.cursor)
	}

	// Move to end
	endMsg := tea.KeyMsg{Type: tea.KeyEnd}
	si.Update(endMsg)
	if si.cursor != len(si.value) {
		t.Errorf("Expected cursor at %d after end, got %d", len(si.value), si.cursor)
	}
}

func TestSearchInput_Update_Delete(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.SetValue("test")
	si.cursor = 0

	// Delete at beginning
	deleteMsg := tea.KeyMsg{Type: tea.KeyDelete}
	si.Update(deleteMsg)

	if si.value != "est" {
		t.Errorf("Expected value %q after delete, got %q", "est", si.value)
	}
}

func TestSearchInput_Update_InsertMiddle(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.SetValue("test")
	si.cursor = 2 // Position between 'e' and 's'

	// Insert 'X'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")}
	si.Update(msg)

	if si.value != "teXst" {
		t.Errorf("Expected value %q after insert, got %q", "teXst", si.value)
	}

	if si.cursor != 3 {
		t.Errorf("Expected cursor at 3 after insert, got %d", si.cursor)
	}
}

func TestSearchInput_Update_Escape(t *testing.T) {
	si := NewSearchInput()
	si.Activate()
	si.SetValue("test")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	si.Update(msg)

	if si.active {
		t.Error("Expected search input to be inactive after Escape")
	}
}

func TestSearchInput_Update_WhenInactive(t *testing.T) {
	si := NewSearchInput()
	// Don't activate

	originalValue := ""
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	si.Update(msg)

	if si.value != originalValue {
		t.Error("Expected no change when inactive")
	}
}

func TestSearchInput_SetSize(t *testing.T) {
	si := NewSearchInput()

	si.SetSize(100, 1)

	if si.width != 100 {
		t.Errorf("Expected width 100, got %d", si.width)
	}

	if si.height != 1 {
		t.Errorf("Expected height 1, got %d", si.height)
	}
}

func TestSearchInput_View(t *testing.T) {
	si := NewSearchInput()
	si.SetSize(50, 1)

	// Test inactive view
	view := si.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Test active view
	si.Activate()
	si.SetValue("test search")
	activeView := si.View()

	if activeView == "" {
		t.Error("Active view should not be empty")
	}

	// Views should be different when active vs inactive
	if view == activeView {
		t.Error("Active and inactive views should differ")
	}
}

func TestSearchInput_IsActive(t *testing.T) {
	si := NewSearchInput()

	if si.IsActive() {
		t.Error("Expected IsActive() to return false initially")
	}

	si.Activate()

	if !si.IsActive() {
		t.Error("Expected IsActive() to return true after Activate()")
	}

	si.Deactivate()

	if si.IsActive() {
		t.Error("Expected IsActive() to return false after Deactivate()")
	}
}

func TestSearchInput_Focus(t *testing.T) {
	si := NewSearchInput()

	// Focus should activate
	si.Focus()
	if !si.active {
		t.Error("Expected Focus() to activate search input")
	}

	// Blur should deactivate
	si.Blur()
	if si.active {
		t.Error("Expected Blur() to deactivate search input")
	}
}

func TestSearchInput_PlaceholderText(t *testing.T) {
	si := NewSearchInput()
	si.SetSize(50, 1)

	// When empty and inactive, should show something (possibly placeholder)
	view := si.View()
	if view == "" {
		t.Error("View should not be empty even when no value is set")
	}
}

func TestSearchInput_SetPlaceholder(t *testing.T) {
	si := NewSearchInput()
	customPlaceholder := "Type to search..."

	si.SetPlaceholder(customPlaceholder)
	si.SetSize(50, 1)

	view := si.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestSearchInput_MoveCursorToEnd(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test value")
	si.cursor = 0

	si.MoveCursorToEnd()

	expectedCursor := len([]rune("test value"))
	if si.cursor != expectedCursor {
		t.Errorf("Expected cursor at %d, got %d", expectedCursor, si.cursor)
	}
}

func TestSearchInput_MoveCursorToStart(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test value")
	si.cursor = 5

	si.MoveCursorToStart()

	if si.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", si.cursor)
	}
}

func TestSearchInput_Insert(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test")
	si.cursor = 4

	si.Insert(" value")

	if si.value != "test value" {
		t.Errorf("Expected value 'test value', got %q", si.value)
	}

	expectedCursor := len([]rune("test value"))
	if si.cursor != expectedCursor {
		t.Errorf("Expected cursor at %d, got %d", expectedCursor, si.cursor)
	}
}

func TestSearchInput_Insert_AtBeginning(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("world")
	si.cursor = 0

	si.Insert("hello ")

	if si.value != "hello world" {
		t.Errorf("Expected value 'hello world', got %q", si.value)
	}
}

func TestSearchInput_DeleteWord(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("hello world test")
	si.cursor = len([]rune("hello world test"))

	si.DeleteWord()

	if si.value != "hello world " {
		t.Errorf("Expected value 'hello world ', got %q", si.value)
	}
}

func TestSearchInput_DeleteWord_AtBeginning(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test")
	si.cursor = 0

	si.DeleteWord()

	if si.value != "test" {
		t.Errorf("Expected value unchanged, got %q", si.value)
	}
}

func TestSearchInput_DeleteWord_SingleWord(t *testing.T) {
	si := NewSearchInput()
	si.SetValue("test")
	si.cursor = len([]rune("test"))

	si.DeleteWord()

	if si.value != "" {
		t.Errorf("Expected empty value, got %q", si.value)
	}
}

func TestSearchInput_SetValue_CursorBeyondLength(t *testing.T) {
	si := NewSearchInput()
	si.cursor = 100

	si.SetValue("short")

	if si.cursor > len(si.value) {
		t.Errorf("Cursor should be adjusted to length, got %d", si.cursor)
	}
}

func TestSearchInput_View_WithCursor(t *testing.T) {
	si := NewSearchInput()
	si.SetSize(80, 1)
	si.Activate()
	si.SetValue("test")
	si.cursor = 2

	view := si.View()
	if view == "" {
		t.Error("View should not be empty with cursor in middle")
	}
}
