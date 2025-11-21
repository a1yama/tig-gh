package models

import "time"

// LeadTimeMetrics はリードタイムに関する統計データを表す
type LeadTimeMetrics struct {
	Overall               LeadTimeStat                               `json:"overall"`
	ByRepository          map[string]LeadTimeStat                    `json:"by_repository"`
	Trend                 []TrendPoint                               `json:"trend"`
	PhaseBreakdown        ReviewPhaseMetrics                         `json:"phase_breakdown"` // 新規追加
	StagnantPRs           StagnantPRMetrics                          `json:"stagnant_prs"`    // 新規追加
	Alerts                AlertMetrics                               `json:"alerts"`          // 新規追加
	ByDayOfWeek           map[time.Weekday]DayOfWeekStats            `json:"by_day_of_week"`
	ByRepositoryDayOfWeek map[string]map[time.Weekday]DayOfWeekStats `json:"by_repository_day_of_week"`
	WeeklyComparison      WeeklyComparison                           `json:"weekly_comparison"`
	ByRepositoryWeekly    map[string]WeeklyComparison                `json:"by_repository_weekly"`
}

// LeadTimeStat は単一リポジトリまたは全体の統計値
type LeadTimeStat struct {
	Average time.Duration `json:"average"`
	Median  time.Duration `json:"median"`
	Count   int           `json:"count"`
}

// TrendPoint は期間ごとの平均リードタイムを表す
type TrendPoint struct {
	Period          string        `json:"period"`
	AverageLeadTime time.Duration `json:"average_lead_time"`
	PRCount         int           `json:"pr_count"`
}

// ReviewPhaseMetrics はレビュー各フェーズの平均時間を表す
type ReviewPhaseMetrics struct {
	CreatedToFirstReview  time.Duration `json:"created_to_first_review"`  // PR作成→最初のレビュー
	FirstReviewToApproval time.Duration `json:"first_review_to_approval"` // 最初のレビュー→承認
	ApprovalToMerge       time.Duration `json:"approval_to_merge"`        // 承認→マージ
	TotalLeadTime         time.Duration `json:"total_lead_time"`          // 全体のリードタイム
}

// StagnantPRInfo は最も古い滞留PRの情報
type StagnantPRInfo struct {
	Repository string        `json:"repository"` // リポジトリ名（owner/repo形式）
	Number     int           `json:"number"`     // PR番号
	Title      string        `json:"title"`      // PRタイトル
	Age        time.Duration `json:"age"`        // 経過時間
}

// StagnantPRMetrics は滞留PR（長期間オープン）の統計
type StagnantPRMetrics struct {
	Threshold      time.Duration    `json:"threshold"`       // 滞留判定のしきい値（例: 72時間）
	TotalStagnant  int              `json:"total_stagnant"`  // 滞留PR総数
	AverageAge     time.Duration    `json:"average_age"`     // 滞留PRの平均経過時間
	LongestWaiting []StagnantPRInfo `json:"longest_waiting"` // 最も古い滞留PR一覧
}

// AlertType はアラートの種類を表す
type AlertType string

const (
	AlertTypeAwaitingReview    AlertType = "awaiting_review"     // レビュー待ち
	AlertTypeAwaitingMerge     AlertType = "awaiting_merge"      // マージ待ち
	AlertTypeLeadTimeIncreased AlertType = "lead_time_increased" // リードタイム増加
)

// Alert は個別のアラート情報
type Alert struct {
	Type     AlertType `json:"type"`     // アラート種別
	Message  string    `json:"message"`  // アラートメッセージ
	Count    int       `json:"count"`    // 該当するPR数（該当する場合）
	Severity string    `json:"severity"` // 重要度（"warning", "critical"など）
}

// AlertMetrics はアラート情報の集合
type AlertMetrics struct {
	Alerts []Alert `json:"alerts"` // アラートのリスト
}

// DayOfWeekStats は曜日ごとのマージ/レビュー件数
type DayOfWeekStats struct {
	ReviewCount int `json:"review_count"`
	MergeCount  int `json:"merge_count"`
}

// WeeklyStats は週次のレビュー/マージ件数
type WeeklyStats struct {
	ReviewCount int `json:"review_count"`
	MergeCount  int `json:"merge_count"`
}

// WeeklyComparison は今週と先週の比較を表す
type WeeklyComparison struct {
	ThisWeek            WeeklyStats `json:"this_week"`
	LastWeek            WeeklyStats `json:"last_week"`
	ReviewChangePercent float64     `json:"review_change_percent"`
	MergeChangePercent  float64     `json:"merge_change_percent"`
}

// MetricsProgress はメトリクス収集の進捗状況を表す
type MetricsProgress struct {
	TotalRepos     int    `json:"total_repos"`     // 総リポジトリ数
	ProcessedRepos int    `json:"processed_repos"` // 処理済みリポジトリ数
	CurrentRepo    string `json:"current_repo"`    // 現在処理中のリポジトリ
}
