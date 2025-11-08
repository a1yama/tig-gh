package keybindings

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyBinding はキーバインディングの定義を表す
type KeyBinding struct {
	// Keys はこのアクションにバインドされているキーのリスト
	Keys []string

	// Action はこのキーバインディングのアクション名
	Action string

	// Description はこのキーバインディングの説明
	Description string

	// Category はこのキーバインディングのカテゴリ（グローバル、ビュー固有など）
	Category string
}

// KeyBindings はキーバインディングの管理を行う
type KeyBindings struct {
	bindings map[string]KeyBinding
}

// NewKeyBindings は新しいKeyBindingsを作成する
func NewKeyBindings() *KeyBindings {
	return &KeyBindings{
		bindings: make(map[string]KeyBinding),
	}
}

// DefaultKeyBindings はデフォルトのキーバインディングを返す
func DefaultKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	// グローバルキーバインディング
	globalBindings := []KeyBinding{
		{
			Keys:        []string{"q"},
			Action:      "quit",
			Description: "終了 / 前の画面に戻る",
			Category:    "global",
		},
		{
			Keys:        []string{"?"},
			Action:      "help",
			Description: "ヘルプ表示",
			Category:    "global",
		},
		{
			Keys:        []string{"r"},
			Action:      "refresh",
			Description: "リフレッシュ",
			Category:    "global",
		},
		{
			Keys:        []string{"/"},
			Action:      "search",
			Description: "検索",
			Category:    "global",
		},
	}

	// ナビゲーションキーバインディング
	navigationBindings := []KeyBinding{
		{
			Keys:        []string{"j", "down"},
			Action:      "down",
			Description: "下に移動",
			Category:    "navigation",
		},
		{
			Keys:        []string{"k", "up"},
			Action:      "up",
			Description: "上に移動",
			Category:    "navigation",
		},
		{
			Keys:        []string{"g"},
			Action:      "first",
			Description: "先頭に移動",
			Category:    "navigation",
		},
		{
			Keys:        []string{"G"},
			Action:      "last",
			Description: "末尾に移動",
			Category:    "navigation",
		},
		{
			Keys:        []string{"ctrl+d"},
			Action:      "page_down",
			Description: "半ページ下",
			Category:    "navigation",
		},
		{
			Keys:        []string{"ctrl+u"},
			Action:      "page_up",
			Description: "半ページ上",
			Category:    "navigation",
		},
		{
			Keys:        []string{"enter"},
			Action:      "select",
			Description: "選択 / 詳細表示",
			Category:    "navigation",
		},
	}

	// アクションキーバインディング
	actionBindings := []KeyBinding{
		{
			Keys:        []string{"o"},
			Action:      "open",
			Description: "ブラウザで開く",
			Category:    "action",
		},
		{
			Keys:        []string{"f"},
			Action:      "filter",
			Description: "フィルタ",
			Category:    "action",
		},
	}

	// ビュー切り替えキーバインディング
	viewSwitchBindings := []KeyBinding{
		{
			Keys:        []string{"i"},
			Action:      "view_1",
			Description: "Issues ビュー",
			Category:    "view",
		},
		{
			Keys:        []string{"p"},
			Action:      "view_2",
			Description: "Pull Requests ビュー",
			Category:    "view",
		},
		{
			Keys:        []string{"c"},
			Action:      "view_3",
			Description: "Commits ビュー",
			Category:    "view",
		},
		{
			Keys:        []string{"4"},
			Action:      "view_4",
			Description: "ビュー4（Notifications）",
			Category:    "view",
		},
		{
			Keys:        []string{"5"},
			Action:      "view_5",
			Description: "ビュー5",
			Category:    "view",
		},
		{
			Keys:        []string{"6"},
			Action:      "view_6",
			Description: "ビュー6",
			Category:    "view",
		},
		{
			Keys:        []string{"7"},
			Action:      "view_7",
			Description: "ビュー7",
			Category:    "view",
		},
		{
			Keys:        []string{"8"},
			Action:      "view_8",
			Description: "ビュー8",
			Category:    "view",
		},
		{
			Keys:        []string{"9"},
			Action:      "view_9",
			Description: "ビュー9",
			Category:    "view",
		},
	}

	// すべてのバインディングを追加
	allBindings := append(globalBindings, navigationBindings...)
	allBindings = append(allBindings, actionBindings...)
	allBindings = append(allBindings, viewSwitchBindings...)

	for _, binding := range allBindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// GetBinding は指定されたアクションのキーバインディングを取得する
func (kb *KeyBindings) GetBinding(action string) (KeyBinding, bool) {
	binding, ok := kb.bindings[action]
	return binding, ok
}

// LoadCustom はカスタムキーバインディングを読み込む
// 既存のバインディングを上書きするか、新しいバインディングを追加する
func (kb *KeyBindings) LoadCustom(custom map[string]string) error {
	for action, keys := range custom {
		if keys == "" {
			return fmt.Errorf("empty key binding for action %q", action)
		}

		// 既存のバインディングを取得または新規作成
		binding, exists := kb.bindings[action]
		if !exists {
			binding = KeyBinding{
				Action:      action,
				Description: action,
				Category:    "custom",
			}
		}

		// キーを更新
		binding.Keys = []string{keys}
		kb.bindings[action] = binding
	}

	return nil
}

// MatchKey は与えられたキーメッセージにマッチするアクションを返す
func (kb *KeyBindings) MatchKey(msg tea.Msg) (string, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", false
	}

	keyStr := keyToString(keyMsg)

	// すべてのバインディングをチェック
	for action, binding := range kb.bindings {
		for _, key := range binding.Keys {
			if key == keyStr {
				return action, true
			}
		}
	}

	return "", false
}

// IsViewSwitch はビュー切り替えキー（1-9）かどうかを判定し、ビュー番号を返す
func (kb *KeyBindings) IsViewSwitch(msg tea.Msg) (int, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return 0, false
	}

	if keyMsg.Type != tea.KeyRunes || len(keyMsg.Runes) != 1 {
		return 0, false
	}

	r := keyMsg.Runes[0]
	if r >= '1' && r <= '9' {
		return int(r - '0'), true
	}

	return 0, false
}

// GetHelp はヘルプ表示用のキーバインディング一覧を返す
func (kb *KeyBindings) GetHelp() []KeyBinding {
	help := make([]KeyBinding, 0, len(kb.bindings))
	for _, binding := range kb.bindings {
		help = append(help, binding)
	}
	return help
}

// keyToString はキーメッセージを文字列に変換する
func keyToString(key tea.KeyMsg) string {
	switch key.Type {
	case tea.KeyRunes:
		if len(key.Runes) > 0 {
			return string(key.Runes)
		}
	case tea.KeyEnter:
		return "enter"
	case tea.KeySpace:
		return "space"
	case tea.KeyTab:
		return "tab"
	case tea.KeyBackspace:
		return "backspace"
	case tea.KeyDelete:
		return "delete"
	case tea.KeyEsc:
		return "esc"
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeyLeft:
		return "left"
	case tea.KeyRight:
		return "right"
	case tea.KeyHome:
		return "home"
	case tea.KeyEnd:
		return "end"
	case tea.KeyPgUp:
		return "pgup"
	case tea.KeyPgDown:
		return "pgdown"
	case tea.KeyCtrlA:
		return "ctrl+a"
	case tea.KeyCtrlB:
		return "ctrl+b"
	case tea.KeyCtrlC:
		return "ctrl+c"
	case tea.KeyCtrlD:
		return "ctrl+d"
	case tea.KeyCtrlE:
		return "ctrl+e"
	case tea.KeyCtrlF:
		return "ctrl+f"
	case tea.KeyCtrlG:
		return "ctrl+g"
	case tea.KeyCtrlH:
		return "ctrl+h"
	case tea.KeyCtrlJ:
		return "ctrl+j"
	case tea.KeyCtrlK:
		return "ctrl+k"
	case tea.KeyCtrlL:
		return "ctrl+l"
	case tea.KeyCtrlN:
		return "ctrl+n"
	case tea.KeyCtrlO:
		return "ctrl+o"
	case tea.KeyCtrlP:
		return "ctrl+p"
	case tea.KeyCtrlQ:
		return "ctrl+q"
	case tea.KeyCtrlR:
		return "ctrl+r"
	case tea.KeyCtrlS:
		return "ctrl+s"
	case tea.KeyCtrlT:
		return "ctrl+t"
	case tea.KeyCtrlU:
		return "ctrl+u"
	case tea.KeyCtrlV:
		return "ctrl+v"
	case tea.KeyCtrlW:
		return "ctrl+w"
	case tea.KeyCtrlX:
		return "ctrl+x"
	case tea.KeyCtrlY:
		return "ctrl+y"
	case tea.KeyCtrlZ:
		return "ctrl+z"
	}

	return ""
}
