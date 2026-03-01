package export

import "testing"

func TestFormatChangePercent(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "正の値", value: 25.0, want: "+25.0%"},
		{name: "負の値", value: -10.5, want: "-10.5%"},
		{name: "ゼロ", value: 0.0, want: "+0.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatChangePercent(tt.value)
			if got != tt.want {
				t.Errorf("formatChangePercent(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
