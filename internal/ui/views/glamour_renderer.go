package views

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

// newMarkdownRenderer creates a glamour renderer without using auto style.
// Auto style triggers OSC background queries that hang in some terminals
// (e.g. WezTerm). Instead we default to the dark theme and only honour
// user overrides via GLAMOUR_STYLE.
func newMarkdownRenderer(wordWrap int) *glamour.TermRenderer {
	style := strings.ToLower(os.Getenv("GLAMOUR_STYLE"))
	if style == "" {
		style = "dark"
	}

	opts := []glamour.TermRendererOption{
		glamour.WithWordWrap(wordWrap),
	}

	switch style {
	case "auto":
		opts = append(opts, glamour.WithAutoStyle())
	default:
		opts = append(opts, glamour.WithStylePath(style))
	}

	renderer, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		// Fall back to a minimal renderer without extra styling.
		fallback, _ := glamour.NewTermRenderer(glamour.WithWordWrap(wordWrap))
		return fallback
	}
	return renderer
}
