package export

import "fmt"

func formatChangePercent(value float64) string {
	return fmt.Sprintf("%+.1f%%", value)
}
