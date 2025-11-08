package keybindings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyBindings(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		wantKeys []string
	}{
		{
			name:     "quit key",
			action:   "quit",
			wantKeys: []string{"q"},
		},
		{
			name:     "help key",
			action:   "help",
			wantKeys: []string{"?"},
		},
		{
			name:     "refresh key",
			action:   "refresh",
			wantKeys: []string{"r"},
		},
		{
			name:     "search key",
			action:   "search",
			wantKeys: []string{"/"},
		},
		{
			name:     "navigation up",
			action:   "up",
			wantKeys: []string{"k", "up"},
		},
		{
			name:     "navigation down",
			action:   "down",
			wantKeys: []string{"j", "down"},
		},
		{
			name:     "select/enter",
			action:   "select",
			wantKeys: []string{"enter"},
		},
	}

	kb := DefaultKeyBindings()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding, ok := kb.bindings[tt.action]
			if !ok {
				t.Errorf("action %q not found in default keybindings", tt.action)
				return
			}

			if binding.Action != tt.action {
				t.Errorf("binding.Action = %v, want %v", binding.Action, tt.action)
			}

			if len(binding.Keys) != len(tt.wantKeys) {
				t.Errorf("binding.Keys length = %v, want %v", len(binding.Keys), len(tt.wantKeys))
				return
			}

			for i, key := range tt.wantKeys {
				if binding.Keys[i] != key {
					t.Errorf("binding.Keys[%d] = %v, want %v", i, binding.Keys[i], key)
				}
			}
		})
	}
}

func TestKeyBindings_GetBinding(t *testing.T) {
	kb := DefaultKeyBindings()

	tests := []struct {
		name       string
		action     string
		wantAction string
		wantFound  bool
	}{
		{
			name:       "existing action quit",
			action:     "quit",
			wantAction: "quit",
			wantFound:  true,
		},
		{
			name:       "existing action help",
			action:     "help",
			wantAction: "help",
			wantFound:  true,
		},
		{
			name:       "non-existing action",
			action:     "nonexistent",
			wantAction: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding, found := kb.GetBinding(tt.action)
			if found != tt.wantFound {
				t.Errorf("GetBinding(%q) found = %v, want %v", tt.action, found, tt.wantFound)
				return
			}

			if found && binding.Action != tt.wantAction {
				t.Errorf("GetBinding(%q).Action = %v, want %v", tt.action, binding.Action, tt.wantAction)
			}
		})
	}
}

func TestKeyBindings_LoadCustom(t *testing.T) {
	kb := DefaultKeyBindings()

	custom := map[string]string{
		"quit":    "x",
		"refresh": "ctrl+r",
		"custom":  "c",
	}

	if err := kb.LoadCustom(custom); err != nil {
		t.Fatalf("LoadCustom() error = %v", err)
	}

	// quit キーが上書きされているか確認
	binding, found := kb.GetBinding("quit")
	if !found {
		t.Error("quit binding not found after LoadCustom")
		return
	}
	if len(binding.Keys) != 1 || binding.Keys[0] != "x" {
		t.Errorf("quit binding.Keys = %v, want [x]", binding.Keys)
	}

	// refresh キーが上書きされているか確認
	binding, found = kb.GetBinding("refresh")
	if !found {
		t.Error("refresh binding not found after LoadCustom")
		return
	}
	if len(binding.Keys) != 1 || binding.Keys[0] != "ctrl+r" {
		t.Errorf("refresh binding.Keys = %v, want [ctrl+r]", binding.Keys)
	}

	// custom キーが追加されているか確認
	binding, found = kb.GetBinding("custom")
	if !found {
		t.Error("custom binding not found after LoadCustom")
		return
	}
	if len(binding.Keys) != 1 || binding.Keys[0] != "c" {
		t.Errorf("custom binding.Keys = %v, want [c]", binding.Keys)
	}
}

func TestKeyBindings_Validate(t *testing.T) {
	tests := []struct {
		name      string
		custom    map[string]string
		wantError bool
	}{
		{
			name: "valid custom bindings",
			custom: map[string]string{
				"quit":    "q",
				"refresh": "r",
			},
			wantError: false,
		},
		{
			name: "empty key value",
			custom: map[string]string{
				"quit": "",
			},
			wantError: true,
		},
		{
			name: "valid with modifiers",
			custom: map[string]string{
				"quit":    "ctrl+q",
				"refresh": "shift+r",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := DefaultKeyBindings()
			err := kb.LoadCustom(tt.custom)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadCustom() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestKeyBindings_MatchKey(t *testing.T) {
	kb := DefaultKeyBindings()

	tests := []struct {
		name       string
		msg        tea.Msg
		wantAction string
		wantMatch  bool
	}{
		{
			name:       "match quit key",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantAction: "quit",
			wantMatch:  true,
		},
		{
			name:       "match help key",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			wantAction: "help",
			wantMatch:  true,
		},
		{
			name:       "match enter key",
			msg:        tea.KeyMsg{Type: tea.KeyEnter},
			wantAction: "select",
			wantMatch:  true,
		},
		{
			name:       "no match",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}},
			wantAction: "",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, match := kb.MatchKey(tt.msg)
			if match != tt.wantMatch {
				t.Errorf("MatchKey() match = %v, want %v", match, tt.wantMatch)
				return
			}
			if match && action != tt.wantAction {
				t.Errorf("MatchKey() action = %v, want %v", action, tt.wantAction)
			}
		})
	}
}

func TestKeyBindings_GetHelp(t *testing.T) {
	kb := DefaultKeyBindings()

	help := kb.GetHelp()
	if len(help) == 0 {
		t.Error("GetHelp() returned empty slice")
	}

	// 基本的なキーがヘルプに含まれているか確認
	foundQuit := false
	foundHelp := false
	for _, h := range help {
		if h.Action == "quit" {
			foundQuit = true
		}
		if h.Action == "help" {
			foundHelp = true
		}
	}

	if !foundQuit {
		t.Error("quit action not found in help")
	}
	if !foundHelp {
		t.Error("help action not found in help")
	}
}

func TestViewNumbers(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantView int
		wantOk   bool
	}{
		{
			name:     "view 1",
			key:      "1",
			wantView: 1,
			wantOk:   true,
		},
		{
			name:     "view 5",
			key:      "5",
			wantView: 5,
			wantOk:   true,
		},
		{
			name:     "view 9",
			key:      "9",
			wantView: 9,
			wantOk:   true,
		},
		{
			name:     "invalid view 0",
			key:      "0",
			wantView: 0,
			wantOk:   false,
		},
		{
			name:     "invalid non-number",
			key:      "a",
			wantView: 0,
			wantOk:   false,
		},
	}

	kb := DefaultKeyBindings()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			view, ok := kb.IsViewSwitch(msg)
			if ok != tt.wantOk {
				t.Errorf("IsViewSwitch(%q) ok = %v, want %v", tt.key, ok, tt.wantOk)
				return
			}
			if ok && view != tt.wantView {
				t.Errorf("IsViewSwitch(%q) view = %v, want %v", tt.key, view, tt.wantView)
			}
		})
	}
}
