package metricsformat

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "正常系: 日+時間+分",
			duration: 2*24*time.Hour + 3*time.Hour + 15*time.Minute,
			want:     "2d 3h 15m",
		},
		{
			name:     "正常系: 日+時間のみ",
			duration: 1*24*time.Hour + 12*time.Hour,
			want:     "1d 12h",
		},
		{
			name:     "正常系: 時間のみ",
			duration: 5 * time.Hour,
			want:     "5h",
		},
		{
			name:     "正常系: 分のみ",
			duration: 30 * time.Minute,
			want:     "30m",
		},
		{
			name:     "正常系: 日のみ",
			duration: 3 * 24 * time.Hour,
			want:     "3d",
		},
		{
			name:     "正常系: 時間+分",
			duration: 2*time.Hour + 45*time.Minute,
			want:     "2h 45m",
		},
		{
			name:     "境界値: ゼロ",
			duration: 0,
			want:     "-",
		},
		{
			name:     "境界値: 負の値",
			duration: -1 * time.Hour,
			want:     "-",
		},
		{
			name:     "境界値: 30秒はRoundで1分に繰り上がる",
			duration: 30 * time.Second,
			want:     "1m",
		},
		{
			name:     "境界値: 29秒はRoundで0秒に丸まる",
			duration: 29 * time.Second,
			want:     "0s",
		},
		{
			name:     "境界値: ちょうど1分",
			duration: 1 * time.Minute,
			want:     "1m",
		},
		{
			name:     "正常系: 分に丸められる端数",
			duration: 2*time.Hour + 30*time.Minute + 29*time.Second,
			want:     "2h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			got := FormatDuration(tt.duration)

			// Then
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestShortWeekday(t *testing.T) {
	tests := []struct {
		day  time.Weekday
		want string
	}{
		{time.Monday, "Mon"},
		{time.Tuesday, "Tue"},
		{time.Wednesday, "Wed"},
		{time.Thursday, "Thu"},
		{time.Friday, "Fri"},
		{time.Saturday, "Sat"},
		{time.Sunday, "Sun"},
	}

	for _, tt := range tests {
		t.Run(tt.day.String(), func(t *testing.T) {
			got := ShortWeekday(tt.day)
			if got != tt.want {
				t.Errorf("ShortWeekday(%v) = %q, want %q", tt.day, got, tt.want)
			}
		})
	}
}
