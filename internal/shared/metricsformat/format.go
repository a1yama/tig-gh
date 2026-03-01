package metricsformat

import (
	"fmt"
	"strings"
	"time"
)

// MaxQualityIssuesToDisplay は品質問題一覧の最大表示件数。
const MaxQualityIssuesToDisplay = 5

// WeekdayDisplayOrder は月曜〜日曜の全曜日を表示順で並べたスライス。
var WeekdayDisplayOrder = []time.Weekday{
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
	time.Saturday,
	time.Sunday,
}

// WeekdaysOnly は平日（月〜金）のみを表示順で並べたスライス。
var WeekdaysOnly = []time.Weekday{
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
}

// FormatDuration は time.Duration を "2d 3h 15m" のような人間が読みやすい形式に変換する。
// 分未満は Round で丸める。0以下の場合は "-" を返す。
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}

	d = d.Round(time.Minute)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	return strings.Join(parts, " ")
}

// ShortWeekday は曜日名を3文字の省略形に変換する（例: Monday → Mon）。
func ShortWeekday(day time.Weekday) string {
	name := day.String()
	if len(name) <= 3 {
		return name
	}
	return name[:3]
}
