package keybindings

import "testing"

func TestGetGlobalKeyBindings(t *testing.T) {
	kb := DefaultKeyBindings()
	globalBindings := kb.GetGlobalKeyBindings()

	if len(globalBindings) == 0 {
		t.Error("GetGlobalKeyBindings() returned empty slice")
	}

	// 必須のグローバルアクションが含まれているか確認
	requiredActions := []string{
		ActionQuit,
		ActionHelp,
		ActionRefresh,
		ActionSearch,
		ActionFilter,
	}

	foundActions := make(map[string]bool)
	for _, binding := range globalBindings {
		foundActions[binding.Action] = true
	}

	for _, action := range requiredActions {
		if !foundActions[action] {
			t.Errorf("required global action %q not found", action)
		}
	}
}

func TestGetNavigationKeyBindings(t *testing.T) {
	kb := DefaultKeyBindings()
	navBindings := kb.GetNavigationKeyBindings()

	if len(navBindings) == 0 {
		t.Error("GetNavigationKeyBindings() returned empty slice")
	}

	// 必須のナビゲーションアクションが含まれているか確認
	requiredActions := []string{
		ActionUp,
		ActionDown,
		ActionFirst,
		ActionLast,
		ActionPageUp,
		ActionPageDown,
		ActionSelect,
	}

	foundActions := make(map[string]bool)
	for _, binding := range navBindings {
		foundActions[binding.Action] = true
	}

	for _, action := range requiredActions {
		if !foundActions[action] {
			t.Errorf("required navigation action %q not found", action)
		}
	}
}

func TestIsGlobalAction(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		wantGlobal bool
	}{
		{
			name:       "quit is global",
			action:     ActionQuit,
			wantGlobal: true,
		},
		{
			name:       "help is global",
			action:     ActionHelp,
			wantGlobal: true,
		},
		{
			name:       "refresh is global",
			action:     ActionRefresh,
			wantGlobal: true,
		},
		{
			name:       "search is global",
			action:     ActionSearch,
			wantGlobal: true,
		},
		{
			name:       "filter is global",
			action:     ActionFilter,
			wantGlobal: true,
		},
		{
			name:       "up is not global",
			action:     ActionUp,
			wantGlobal: false,
		},
		{
			name:       "new_issue is not global",
			action:     ActionNewIssue,
			wantGlobal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGlobalAction(tt.action); got != tt.wantGlobal {
				t.Errorf("IsGlobalAction(%q) = %v, want %v", tt.action, got, tt.wantGlobal)
			}
		})
	}
}

func TestActionConstants(t *testing.T) {
	// アクション定数が空でないことを確認
	tests := []struct {
		name   string
		action string
	}{
		{"ActionQuit", ActionQuit},
		{"ActionHelp", ActionHelp},
		{"ActionRefresh", ActionRefresh},
		{"ActionSearch", ActionSearch},
		{"ActionFilter", ActionFilter},
		{"ActionUp", ActionUp},
		{"ActionDown", ActionDown},
		{"ActionFirst", ActionFirst},
		{"ActionLast", ActionLast},
		{"ActionPageUp", ActionPageUp},
		{"ActionPageDown", ActionPageDown},
		{"ActionSelect", ActionSelect},
		{"ActionOpen", ActionOpen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "" {
				t.Errorf("%s is empty", tt.name)
			}
		})
	}
}
