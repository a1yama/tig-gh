package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// カラーパレット
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan
	ColorAccent    = lipgloss.Color("#F59E0B") // Amber

	// State colors
	ColorSuccess = lipgloss.Color("#10B981") // Green
	ColorWarning = lipgloss.Color("#F59E0B") // Amber
	ColorError   = lipgloss.Color("#EF4444") // Red
	ColorInfo    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	ColorForeground = lipgloss.Color("#E5E7EB") // Light gray
	ColorBackground = lipgloss.Color("#111827") // Dark gray
	ColorMuted      = lipgloss.Color("#6B7280") // Medium gray
	ColorBorder     = lipgloss.Color("#374151") // Border gray

	// Issue/PR state colors
	ColorOpen   = lipgloss.Color("#10B981") // Green
	ColorClosed = lipgloss.Color("#EF4444") // Red
	ColorMerged = lipgloss.Color("#7C3AED") // Purple
)

// 基本スタイル
var (
	// テキストスタイル
	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	BoldStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// ヘッダースタイル
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// タイトルスタイル
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Bold(true).
			Padding(0, 1)

	// 選択状態のスタイル
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorPrimary).
			Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// ボーダースタイル
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// ステータスバースタイル
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Background(ColorBackground).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StatusValueStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// ヘルプスタイル
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// エラースタイル
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true).
			Padding(1, 2)

	// 成功スタイル
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// 警告スタイル
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// 情報スタイル
	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)

	// ローディングスタイル
	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)
)

// リストアイテムのスタイル
var (
	// Issue番号
	IssueNumberStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Bold(false)

	// Issueタイトル
	IssueTitleStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	// Issue状態
	IssueOpenStyle = lipgloss.NewStyle().
			Foreground(ColorOpen).
			Bold(true)

	IssueClosedStyle = lipgloss.NewStyle().
				Foreground(ColorClosed).
				Bold(true)

	// ラベル
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorBackground).
			Background(ColorAccent).
			Padding(0, 1).
			MarginRight(1)

	// 日付
	DateStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// 作成者
	AuthorStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)

// GetStateStyle returns the style for the given state
func GetStateStyle(state string) lipgloss.Style {
	switch state {
	case "open":
		return IssueOpenStyle
	case "closed":
		return IssueClosedStyle
	case "merged":
		return lipgloss.NewStyle().Foreground(ColorMerged).Bold(true)
	default:
		return MutedStyle
	}
}

// GetStateBadge returns a styled badge for the given state
func GetStateBadge(state string) string {
	style := GetStateStyle(state)
	switch state {
	case "open":
		return style.Render("● OPEN")
	case "closed":
		return style.Render("● CLOSED")
	case "merged":
		return style.Render("● MERGED")
	default:
		return style.Render("● " + state)
	}
}

// ヘルプテキストのフォーマット
func FormatKeyBinding(key, desc string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		HelpKeyStyle.Render(key),
		HelpDescStyle.Render(": "+desc),
		NormalStyle.Render(" "),
	)
}

// セパレータ
func Separator(width int) string {
	return lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(lipgloss.NewStyle().Width(width).Render("─"))
}
