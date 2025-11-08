package keybindings

// GlobalActions はグローバルに使用できるアクション名を定義する
const (
	// ActionQuit は終了/戻るアクション
	ActionQuit = "quit"

	// ActionHelp はヘルプ表示アクション
	ActionHelp = "help"

	// ActionRefresh はリフレッシュアクション
	ActionRefresh = "refresh"

	// ActionSearch は検索アクション
	ActionSearch = "search"

	// ActionFilter はフィルタアクション
	ActionFilter = "filter"
)

// NavigationActions はナビゲーション関連のアクション名を定義する
const (
	// ActionUp は上に移動アクション
	ActionUp = "up"

	// ActionDown は下に移動アクション
	ActionDown = "down"

	// ActionFirst は先頭に移動アクション
	ActionFirst = "first"

	// ActionLast は末尾に移動アクション
	ActionLast = "last"

	// ActionPageUp は半ページ上アクション
	ActionPageUp = "page_up"

	// ActionPageDown は半ページ下アクション
	ActionPageDown = "page_down"

	// ActionSelect は選択/詳細表示アクション
	ActionSelect = "select"
)

// CommonActions は共通で使用されるアクション名を定義する
const (
	// ActionOpen はブラウザで開くアクション
	ActionOpen = "open"
)

// ViewSwitchActions はビュー切り替え関連のアクション名を定義する
const (
	// ActionView1 はビュー1（Issues）
	ActionView1 = "view_1"

	// ActionView2 はビュー2（Pull Requests）
	ActionView2 = "view_2"

	// ActionView3 はビュー3（Commits）
	ActionView3 = "view_3"

	// ActionView4 はビュー4（Notifications）
	ActionView4 = "view_4"

	// ActionView5 はビュー5
	ActionView5 = "view_5"

	// ActionView6 はビュー6
	ActionView6 = "view_6"

	// ActionView7 はビュー7
	ActionView7 = "view_7"

	// ActionView8 はビュー8
	ActionView8 = "view_8"

	// ActionView9 はビュー9
	ActionView9 = "view_9"
)

// GetGlobalKeyBindings はグローバルキーバインディングのみを返す
func (kb *KeyBindings) GetGlobalKeyBindings() []KeyBinding {
	globalActions := []string{
		ActionQuit,
		ActionHelp,
		ActionRefresh,
		ActionSearch,
		ActionFilter,
	}

	bindings := make([]KeyBinding, 0, len(globalActions))
	for _, action := range globalActions {
		if binding, ok := kb.bindings[action]; ok {
			bindings = append(bindings, binding)
		}
	}

	return bindings
}

// GetNavigationKeyBindings はナビゲーションキーバインディングのみを返す
func (kb *KeyBindings) GetNavigationKeyBindings() []KeyBinding {
	navActions := []string{
		ActionUp,
		ActionDown,
		ActionFirst,
		ActionLast,
		ActionPageUp,
		ActionPageDown,
		ActionSelect,
	}

	bindings := make([]KeyBinding, 0, len(navActions))
	for _, action := range navActions {
		if binding, ok := kb.bindings[action]; ok {
			bindings = append(bindings, binding)
		}
	}

	return bindings
}

// IsGlobalAction は指定されたアクションがグローバルアクションかどうかを判定する
func IsGlobalAction(action string) bool {
	globalActions := []string{
		ActionQuit,
		ActionHelp,
		ActionRefresh,
		ActionSearch,
		ActionFilter,
	}

	for _, a := range globalActions {
		if a == action {
			return true
		}
	}

	return false
}
