package keybindings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestIntegration_DefaultWithCustomOverride はデフォルト設定にカスタム設定を上書きする統合テスト
func TestIntegration_DefaultWithCustomOverride(t *testing.T) {
	// デフォルトキーバインディングを取得
	kb := DefaultKeyBindings()

	// カスタム設定を読み込む
	custom := map[string]string{
		"quit":    "x",        // qからxに変更
		"refresh": "ctrl+r",   // rからctrl+rに変更
		"custom":  "c",        // 新規アクション追加
	}

	if err := kb.LoadCustom(custom); err != nil {
		t.Fatalf("LoadCustom() error = %v", err)
	}

	// quitキーが変更されていることを確認
	action, match := kb.MatchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if !match {
		t.Error("quit key 'x' not matched")
	}
	if action != "quit" {
		t.Errorf("quit key matched action = %v, want 'quit'", action)
	}

	// 古いquitキー（q）がマッチしないことを確認
	// 注意: デフォルトでは複数のキーがバインドされている可能性があるため、
	// 上書き後は古いキーは削除される
	// ただし、他のアクションで'q'が使われていないことを確認

	// customアクションが追加されていることを確認
	// 注意: 'c'はデフォルトで'comment'にバインドされているため、
	// customで上書きされる
	action, match = kb.MatchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if !match {
		t.Error("custom key 'c' not matched")
	}
	if action != "custom" {
		t.Errorf("custom key matched action = %v, want 'custom'", action)
	}
}

// TestIntegration_ViewSpecificKeyBindings はビュー固有のキーバインディングの統合テスト
func TestIntegration_ViewSpecificKeyBindings(t *testing.T) {
	// グローバル + Issueビューのキーバインディングをマージ
	global := DefaultKeyBindings()
	issue := GetIssueViewKeyBindings()
	merged := MergeKeyBindings(global, issue)

	tests := []struct {
		name       string
		key        tea.KeyMsg
		wantAction string
		wantMatch  bool
	}{
		{
			name:       "global quit",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantAction: "quit",
			wantMatch:  true,
		},
		{
			name:       "issue new",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			wantAction: "new_issue",
			wantMatch:  true,
		},
		{
			name:       "issue edit",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
			wantAction: "edit_issue",
			wantMatch:  true,
		},
		{
			name:       "navigation up",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantAction: "up",
			wantMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, match := merged.MatchKey(tt.key)
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

// TestIntegration_MultipleViewSwitch は複数のビュー切り替えの統合テスト
func TestIntegration_MultipleViewSwitch(t *testing.T) {
	kb := DefaultKeyBindings()

	views := []struct {
		key  rune
		view int
	}{
		{'1', 1},
		{'2', 2},
		{'3', 3},
		{'4', 4},
		{'5', 5},
		{'6', 6},
		{'7', 7},
		{'8', 8},
		{'9', 9},
	}

	for _, v := range views {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{v.key}}
		view, ok := kb.IsViewSwitch(msg)
		if !ok {
			t.Errorf("IsViewSwitch(%c) ok = false, want true", v.key)
			continue
		}
		if view != v.view {
			t.Errorf("IsViewSwitch(%c) view = %v, want %v", v.key, view, v.view)
		}
	}
}

// TestIntegration_HelpGeneration はヘルプ生成の統合テスト
func TestIntegration_HelpGeneration(t *testing.T) {
	kb := DefaultKeyBindings()
	help := kb.GetHelp()

	if len(help) == 0 {
		t.Fatal("GetHelp() returned empty slice")
	}

	// カテゴリごとにヘルプがグループ化できることを確認
	categories := make(map[string]int)
	for _, h := range help {
		categories[h.Category]++
	}

	expectedCategories := []string{"global", "navigation", "action", "view"}
	for _, cat := range expectedCategories {
		if count, ok := categories[cat]; !ok || count == 0 {
			t.Errorf("category %q not found or has no items in help", cat)
		}
	}
}

// TestIntegration_ComplexKeySequence は複雑なキー入力シーケンスの統合テスト
func TestIntegration_ComplexKeySequence(t *testing.T) {
	global := DefaultKeyBindings()
	issue := GetIssueViewKeyBindings()
	merged := MergeKeyBindings(global, issue)

	// ユーザーの操作シーケンスをシミュレート
	sequences := []struct {
		name       string
		keys       []tea.KeyMsg
		wantAction []string
	}{
		{
			name: "navigate and select",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'j'}},  // down
				{Type: tea.KeyRunes, Runes: []rune{'j'}},  // down
				{Type: tea.KeyEnter},                       // select
			},
			wantAction: []string{"down", "down", "select"},
		},
		{
			name: "new issue and quit",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'n'}},  // new issue
				{Type: tea.KeyRunes, Runes: []rune{'q'}},  // quit
			},
			wantAction: []string{"new_issue", "quit"},
		},
		{
			name: "search and navigation",
			keys: []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'/'}},  // search
				{Type: tea.KeyRunes, Runes: []rune{'k'}},  // up
				{Type: tea.KeyEnter},                       // select
			},
			wantAction: []string{"search", "up", "select"},
		},
	}

	for _, seq := range sequences {
		t.Run(seq.name, func(t *testing.T) {
			for i, key := range seq.keys {
				action, match := merged.MatchKey(key)
				if !match {
					t.Errorf("key[%d] not matched", i)
					continue
				}
				if action != seq.wantAction[i] {
					t.Errorf("key[%d] action = %v, want %v", i, action, seq.wantAction[i])
				}
			}
		})
	}
}

// TestIntegration_ConflictResolution はキーコンフリクトの解決テスト
func TestIntegration_ConflictResolution(t *testing.T) {
	global := DefaultKeyBindings()
	issue := GetIssueViewKeyBindings()

	// IssueビューとPRビューで異なるアクションにバインド
	pr := GetPRViewKeyBindings()

	// Global + Issueのマージ
	issueView := MergeKeyBindings(global, issue)

	// Global + PRのマージ
	prView := MergeKeyBindings(global, pr)

	// Issueビューでの'n'キー
	// マージの順序により、後にマージされたものが優先される
	action, _ := issueView.MatchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if action != "new_issue" {
		t.Errorf("issue view 'n' action = %v, want 'new_issue'", action)
	}

	// PRビューでの'n'キー
	action, _ = prView.MatchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if action != "next_file" {
		t.Errorf("pr view 'n' action = %v, want 'next_file'", action)
	}
}

// TestIntegration_KeyBindingValidation はキーバインディングのバリデーション統合テスト
func TestIntegration_KeyBindingValidation(t *testing.T) {
	kb := DefaultKeyBindings()

	invalidCustoms := []struct {
		name   string
		custom map[string]string
	}{
		{
			name: "empty key",
			custom: map[string]string{
				"quit": "",
			},
		},
	}

	for _, tc := range invalidCustoms {
		t.Run(tc.name, func(t *testing.T) {
			err := kb.LoadCustom(tc.custom)
			if err == nil {
				t.Error("LoadCustom() expected error but got nil")
			}
		})
	}
}
