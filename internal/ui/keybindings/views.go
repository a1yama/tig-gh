package keybindings

// IssueViewActions はIssueビュー固有のアクション名を定義する
const (
	// ActionNewIssue は新規Issue作成アクション
	ActionNewIssue = "new_issue"

	// ActionEditIssue はIssue編集アクション
	ActionEditIssue = "edit_issue"

	// ActionCloseIssue はIssueクローズ/再オープンアクション
	ActionCloseIssue = "close_issue"

	// ActionCommentIssue はIssueコメントアクション
	ActionCommentIssue = "comment_issue"

	// ActionAssignIssue はIssueアサインアクション
	ActionAssignIssue = "assign_issue"

	// ActionLabelIssue はIssueラベル付けアクション
	ActionLabelIssue = "label_issue"
)

// PRViewActions はPull Requestビュー固有のアクション名を定義する
const (
	// ActionDiff は差分表示アクション
	ActionDiff = "diff"

	// ActionMerge はマージアクション
	ActionMerge = "merge"

	// ActionReview はレビューアクション
	ActionReview = "review"

	// ActionApprove は承認アクション
	ActionApprove = "approve"

	// ActionRequestChanges は変更要求アクション
	ActionRequestChanges = "request_changes"

	// ActionNextFile は次のファイルに移動アクション
	ActionNextFile = "next_file"

	// ActionPrevFile は前のファイルに移動アクション
	ActionPrevFile = "prev_file"
)

// CommitViewActions はCommitビュー固有のアクション名を定義する
const (
	// ActionShowCommit はコミット詳細表示アクション
	ActionShowCommit = "show_commit"

	// ActionCopyHash はコミットハッシュコピーアクション
	ActionCopyHash = "copy_hash"

	// ActionCherryPick はチェリーピックアクション
	ActionCherryPick = "cherry_pick"
)

// NotificationViewActions は通知ビュー固有のアクション名を定義する
const (
	// ActionMarkRead は既読にするアクション
	ActionMarkRead = "mark_read"

	// ActionMarkAllRead はすべて既読にするアクション
	ActionMarkAllRead = "mark_all_read"

	// ActionUnsubscribe は購読解除アクション
	ActionUnsubscribe = "unsubscribe"
)

// MetricsViewActions はメトリクスビュー固有のアクション名を定義する
const (
	// ActionMetricsScrollDown は前方スクロール
	ActionMetricsScrollDown = "metrics_scroll_down"

	// ActionMetricsScrollUp は後方スクロール
	ActionMetricsScrollUp = "metrics_scroll_up"

	// ActionMetricsRefresh はメトリクス再取得
	ActionMetricsRefresh = "metrics_refresh"

	// ActionMetricsBack は前のビューへ戻る操作
	ActionMetricsBack = "metrics_back"
)

// GetIssueViewKeyBindings はIssueビュー固有のキーバインディングを返す
func GetIssueViewKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	issueBindings := []KeyBinding{
		{
			Keys:        []string{"n"},
			Action:      ActionNewIssue,
			Description: "新規Issue作成",
			Category:    "issue",
		},
		{
			Keys:        []string{"e"},
			Action:      ActionEditIssue,
			Description: "Issue編集",
			Category:    "issue",
		},
		{
			Keys:        []string{"c"},
			Action:      ActionCommentIssue,
			Description: "コメント追加",
			Category:    "issue",
		},
		{
			Keys:        []string{"x"},
			Action:      ActionCloseIssue,
			Description: "クローズ/再オープン",
			Category:    "issue",
		},
		{
			Keys:        []string{"a"},
			Action:      ActionAssignIssue,
			Description: "アサイン",
			Category:    "issue",
		},
		{
			Keys:        []string{"l"},
			Action:      ActionLabelIssue,
			Description: "ラベル付け",
			Category:    "issue",
		},
	}

	for _, binding := range issueBindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// GetPRViewKeyBindings はPull Requestビュー固有のキーバインディングを返す
func GetPRViewKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	prBindings := []KeyBinding{
		{
			Keys:        []string{"d"},
			Action:      ActionDiff,
			Description: "差分表示",
			Category:    "pr",
		},
		{
			Keys:        []string{"m"},
			Action:      ActionMerge,
			Description: "マージ",
			Category:    "pr",
		},
		{
			Keys:        []string{"r"},
			Action:      ActionReview,
			Description: "レビュー",
			Category:    "pr",
		},
		{
			Keys:        []string{"A"},
			Action:      ActionApprove,
			Description: "承認",
			Category:    "pr",
		},
		{
			Keys:        []string{"C"},
			Action:      ActionRequestChanges,
			Description: "変更要求",
			Category:    "pr",
		},
		{
			Keys:        []string{"n"},
			Action:      ActionNextFile,
			Description: "次のファイル",
			Category:    "pr",
		},
		{
			Keys:        []string{"p"},
			Action:      ActionPrevFile,
			Description: "前のファイル",
			Category:    "pr",
		},
	}

	for _, binding := range prBindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// GetCommitViewKeyBindings はCommitビュー固有のキーバインディングを返す
func GetCommitViewKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	commitBindings := []KeyBinding{
		{
			Keys:        []string{"enter"},
			Action:      ActionShowCommit,
			Description: "コミット詳細表示",
			Category:    "commit",
		},
		{
			Keys:        []string{"d"},
			Action:      ActionDiff,
			Description: "差分表示",
			Category:    "commit",
		},
		{
			Keys:        []string{"y"},
			Action:      ActionCopyHash,
			Description: "ハッシュコピー",
			Category:    "commit",
		},
		{
			Keys:        []string{"C"},
			Action:      ActionCherryPick,
			Description: "チェリーピック",
			Category:    "commit",
		},
	}

	for _, binding := range commitBindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// GetNotificationViewKeyBindings は通知ビュー固有のキーバインディングを返す
func GetNotificationViewKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	notificationBindings := []KeyBinding{
		{
			Keys:        []string{"d"},
			Action:      ActionMarkRead,
			Description: "既読にする",
			Category:    "notification",
		},
		{
			Keys:        []string{"D"},
			Action:      ActionMarkAllRead,
			Description: "すべて既読にする",
			Category:    "notification",
		},
		{
			Keys:        []string{"u"},
			Action:      ActionUnsubscribe,
			Description: "購読解除",
			Category:    "notification",
		},
	}

	for _, binding := range notificationBindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// GetMetricsViewKeyBindings はメトリクスビュー固有のキーバインディングを返す
func GetMetricsViewKeyBindings() *KeyBindings {
	kb := NewKeyBindings()

	bindings := []KeyBinding{
		{
			Keys:        []string{"j", "down"},
			Action:      ActionMetricsScrollDown,
			Description: "下にスクロール",
			Category:    "metrics",
		},
		{
			Keys:        []string{"k", "up"},
			Action:      ActionMetricsScrollUp,
			Description: "上にスクロール",
			Category:    "metrics",
		},
		{
			Keys:        []string{"r"},
			Action:      ActionMetricsRefresh,
			Description: "メトリクス更新",
			Category:    "metrics",
		},
		{
			Keys:        []string{"q"},
			Action:      ActionMetricsBack,
			Description: "前のビューへ戻る",
			Category:    "metrics",
		},
	}

	for _, binding := range bindings {
		kb.bindings[binding.Action] = binding
	}

	return kb
}

// MergeKeyBindings は複数のKeyBindingsをマージする
func MergeKeyBindings(base *KeyBindings, additional ...*KeyBindings) *KeyBindings {
	merged := NewKeyBindings()

	// ベースのバインディングをコピー
	for action, binding := range base.bindings {
		merged.bindings[action] = binding
	}

	// 追加のバインディングをマージ（上書き）
	for _, kb := range additional {
		for action, binding := range kb.bindings {
			merged.bindings[action] = binding
		}
	}

	return merged
}
