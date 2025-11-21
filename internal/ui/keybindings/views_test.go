package keybindings

import "testing"

func TestGetIssueViewKeyBindings(t *testing.T) {
	kb := GetIssueViewKeyBindings()

	if kb == nil {
		t.Fatal("GetIssueViewKeyBindings() returned nil")
	}

	// 必須のIssueアクションが含まれているか確認
	requiredActions := []string{
		ActionNewIssue,
		ActionEditIssue,
		ActionCommentIssue,
		ActionCloseIssue,
		ActionAssignIssue,
		ActionLabelIssue,
	}

	for _, action := range requiredActions {
		if _, ok := kb.GetBinding(action); !ok {
			t.Errorf("required issue action %q not found", action)
		}
	}
}

func TestGetPRViewKeyBindings(t *testing.T) {
	kb := GetPRViewKeyBindings()

	if kb == nil {
		t.Fatal("GetPRViewKeyBindings() returned nil")
	}

	// 必須のPRアクションが含まれているか確認
	requiredActions := []string{
		ActionDiff,
		ActionMerge,
		ActionReview,
		ActionApprove,
		ActionRequestChanges,
		ActionNextFile,
		ActionPrevFile,
	}

	for _, action := range requiredActions {
		if _, ok := kb.GetBinding(action); !ok {
			t.Errorf("required PR action %q not found", action)
		}
	}
}

func TestGetCommitViewKeyBindings(t *testing.T) {
	kb := GetCommitViewKeyBindings()

	if kb == nil {
		t.Fatal("GetCommitViewKeyBindings() returned nil")
	}

	// 必須のCommitアクションが含まれているか確認
	requiredActions := []string{
		ActionShowCommit,
		ActionDiff,
		ActionCopyHash,
		ActionCherryPick,
	}

	for _, action := range requiredActions {
		if _, ok := kb.GetBinding(action); !ok {
			t.Errorf("required commit action %q not found", action)
		}
	}
}

func TestGetNotificationViewKeyBindings(t *testing.T) {
	kb := GetNotificationViewKeyBindings()

	if kb == nil {
		t.Fatal("GetNotificationViewKeyBindings() returned nil")
	}

	// 必須のNotificationアクションが含まれているか確認
	requiredActions := []string{
		ActionMarkRead,
		ActionMarkAllRead,
		ActionUnsubscribe,
	}

	for _, action := range requiredActions {
		if _, ok := kb.GetBinding(action); !ok {
			t.Errorf("required notification action %q not found", action)
		}
	}
}

func TestGetMetricsViewKeyBindings(t *testing.T) {
	kb := GetMetricsViewKeyBindings()

	if kb == nil {
		t.Fatal("GetMetricsViewKeyBindings() returned nil")
	}

	required := []string{
		ActionMetricsScrollDown,
		ActionMetricsScrollUp,
		ActionMetricsRefresh,
		ActionMetricsBack,
	}

	for _, action := range required {
		if _, ok := kb.GetBinding(action); !ok {
			t.Errorf("required metrics action %q not found", action)
		}
	}
}

func TestMergeKeyBindings(t *testing.T) {
	base := DefaultKeyBindings()
	issue := GetIssueViewKeyBindings()
	pr := GetPRViewKeyBindings()

	merged := MergeKeyBindings(base, issue, pr)

	if merged == nil {
		t.Fatal("MergeKeyBindings() returned nil")
	}

	// ベースのキーバインディングが含まれているか確認
	if _, ok := merged.GetBinding(ActionQuit); !ok {
		t.Error("base action 'quit' not found in merged bindings")
	}

	// Issueのキーバインディングが含まれているか確認
	if _, ok := merged.GetBinding(ActionNewIssue); !ok {
		t.Error("issue action 'new_issue' not found in merged bindings")
	}

	// PRのキーバインディングが含まれているか確認
	if _, ok := merged.GetBinding(ActionMerge); !ok {
		t.Error("PR action 'merge' not found in merged bindings")
	}
}

func TestMergeKeyBindings_Overwrite(t *testing.T) {
	base := DefaultKeyBindings()
	custom := NewKeyBindings()

	// カスタムで'quit'を上書き
	custom.bindings["quit"] = KeyBinding{
		Keys:        []string{"x"},
		Action:      "quit",
		Description: "カスタム終了",
		Category:    "custom",
	}

	merged := MergeKeyBindings(base, custom)

	binding, ok := merged.GetBinding("quit")
	if !ok {
		t.Fatal("quit binding not found in merged bindings")
	}

	// カスタムの値で上書きされているか確認
	if len(binding.Keys) != 1 || binding.Keys[0] != "x" {
		t.Errorf("quit binding not overwritten, got keys = %v", binding.Keys)
	}

	if binding.Description != "カスタム終了" {
		t.Errorf("quit binding description = %q, want 'カスタム終了'", binding.Description)
	}
}

func TestViewActionConstants(t *testing.T) {
	// ビュー固有のアクション定数が空でないことを確認
	tests := []struct {
		name   string
		action string
	}{
		// Issue actions
		{"ActionNewIssue", ActionNewIssue},
		{"ActionEditIssue", ActionEditIssue},
		{"ActionCloseIssue", ActionCloseIssue},
		{"ActionCommentIssue", ActionCommentIssue},
		{"ActionAssignIssue", ActionAssignIssue},
		{"ActionLabelIssue", ActionLabelIssue},

		// PR actions
		{"ActionDiff", ActionDiff},
		{"ActionMerge", ActionMerge},
		{"ActionReview", ActionReview},
		{"ActionApprove", ActionApprove},
		{"ActionRequestChanges", ActionRequestChanges},
		{"ActionNextFile", ActionNextFile},
		{"ActionPrevFile", ActionPrevFile},

		// Commit actions
		{"ActionShowCommit", ActionShowCommit},
		{"ActionCopyHash", ActionCopyHash},
		{"ActionCherryPick", ActionCherryPick},

		// Notification actions
		{"ActionMarkRead", ActionMarkRead},
		{"ActionMarkAllRead", ActionMarkAllRead},
		{"ActionUnsubscribe", ActionUnsubscribe},

		// Metrics actions
		{"ActionMetricsScrollDown", ActionMetricsScrollDown},
		{"ActionMetricsScrollUp", ActionMetricsScrollUp},
		{"ActionMetricsRefresh", ActionMetricsRefresh},
		{"ActionMetricsBack", ActionMetricsBack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action == "" {
				t.Errorf("%s is empty", tt.name)
			}
		})
	}
}

func TestViewKeyBindings_NoConflict(t *testing.T) {
	// ビュー間でアクション名の衝突がないことを確認
	issueKB := GetIssueViewKeyBindings()
	prKB := GetPRViewKeyBindings()
	commitKB := GetCommitViewKeyBindings()
	notifKB := GetNotificationViewKeyBindings()
	metricsKB := GetMetricsViewKeyBindings()

	allActions := make(map[string]string)

	// Issueアクションを追加
	for action := range issueKB.bindings {
		allActions[action] = "issue"
	}

	// PRアクションをチェック（diffは共通なので許可）
	for action := range prKB.bindings {
		if view, exists := allActions[action]; exists && action != ActionDiff {
			t.Errorf("action %q conflicts between %s and pr views", action, view)
		}
		allActions[action] = "pr"
	}

	// Commitアクションをチェック（diffは共通なので許可）
	for action := range commitKB.bindings {
		if view, exists := allActions[action]; exists && action != ActionDiff && action != ActionShowCommit {
			t.Errorf("action %q conflicts between %s and commit views", action, view)
		}
		allActions[action] = "commit"
	}

	// Notificationアクションをチェック
	for action := range notifKB.bindings {
		if view, exists := allActions[action]; exists {
			t.Errorf("action %q conflicts between %s and notification views", action, view)
		}
		allActions[action] = "notification"
	}

	for action := range metricsKB.bindings {
		if view, exists := allActions[action]; exists {
			t.Errorf("action %q conflicts between %s and metrics views", action, view)
		}
		allActions[action] = "metrics"
	}
}
