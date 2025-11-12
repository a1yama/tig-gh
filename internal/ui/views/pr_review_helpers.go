package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/a1yama/tig-gh/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// renderReviewSummary formats review counts into a compact summary string.
func renderReviewSummary(reviews []models.Review) string {
	var summary []string
	reviewCounts := make(map[models.ReviewState]int)

	for _, review := range reviews {
		reviewCounts[review.State]++
	}

	if count := reviewCounts[models.ReviewStateApproved]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")).
			Render(fmt.Sprintf("✓%d", count)))
	}

	if count := reviewCounts[models.ReviewStateChangesRequested]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("✗%d", count)))
	}

	if count := reviewCounts[models.ReviewStatePending]; count > 0 {
		summary = append(summary, lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Render(fmt.Sprintf("?%d", count)))
	}

	if len(summary) == 0 {
		return styles.MutedStyle.Render("No reviews")
	}

	return strings.Join(summary, " ")
}

// firstReviewSubmittedAt returns the earliest non-pending review submission time.
func firstReviewSubmittedAt(reviews []models.Review) *time.Time {
	var earliest *time.Time
	for _, review := range reviews {
		if review.SubmittedAt.IsZero() {
			continue
		}
		if review.State == models.ReviewStatePending {
			continue
		}
		reviewTime := review.SubmittedAt
		if earliest == nil || reviewTime.Before(*earliest) {
			earliest = &reviewTime
		}
	}
	return earliest
}

// firstApprovalSubmittedAt returns the earliest approval submission time.
func firstApprovalSubmittedAt(reviews []models.Review) *time.Time {
	var earliest *time.Time
	for _, review := range reviews {
		if review.State != models.ReviewStateApproved {
			continue
		}
		if review.SubmittedAt.IsZero() {
			continue
		}
		reviewTime := review.SubmittedAt
		if earliest == nil || reviewTime.Before(*earliest) {
			earliest = &reviewTime
		}
	}
	return earliest
}

// formatDurationBetween formats the duration between two timestamps using formatDurationShort.
func formatDurationBetween(start, end time.Time) string {
	if end.Before(start) {
		start, end = end, start
	}
	return formatDurationShort(end.Sub(start))
}

// formatDurationShort formats durations like "2d 3h" keeping at most 2 units.
func formatDurationShort(d time.Duration) string {
	if d <= 0 {
		return "<1m"
	}
	var parts []string
	days := d / (24 * time.Hour)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
		d -= days * 24 * time.Hour
	}
	hours := d / time.Hour
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		d -= hours * time.Hour
	}
	minutes := d / time.Minute
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		d -= minutes * time.Minute
	}
	if len(parts) == 0 {
		seconds := int((d + time.Second/2) / time.Second)
		if seconds <= 0 {
			seconds = 1
		}
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

// flattenReviews converts review pointers to value slices.
func flattenReviews(reviews []*models.Review) []models.Review {
	if len(reviews) == 0 {
		return nil
	}
	result := make([]models.Review, 0, len(reviews))
	for _, review := range reviews {
		if review == nil {
			continue
		}
		result = append(result, *review)
	}
	return result
}
