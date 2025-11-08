package git

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "HTTPS URL with .git suffix",
			url:       "https://github.com/charmbracelet/bubbletea.git",
			wantOwner: "charmbracelet",
			wantRepo:  "bubbletea",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL without .git suffix",
			url:       "https://github.com/charmbracelet/bubbletea",
			wantOwner: "charmbracelet",
			wantRepo:  "bubbletea",
			wantErr:   false,
		},
		{
			name:      "SSH URL with .git suffix",
			url:       "git@github.com:charmbracelet/bubbletea.git",
			wantOwner: "charmbracelet",
			wantRepo:  "bubbletea",
			wantErr:   false,
		},
		{
			name:      "SSH URL without .git suffix",
			url:       "git@github.com:charmbracelet/bubbletea",
			wantOwner: "charmbracelet",
			wantRepo:  "bubbletea",
			wantErr:   false,
		},
		{
			name:      "URL with whitespace",
			url:       "  https://github.com/owner/repo.git  ",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:    "Invalid URL format",
			url:     "https://gitlab.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "Invalid repository format (missing repo)",
			url:     "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "Invalid repository format (too many parts)",
			url:     "https://github.com/owner/repo/extra",
			wantErr: true,
		},
		{
			name:    "Empty owner",
			url:     "https://github.com//repo",
			wantErr: true,
		},
		{
			name:    "Empty repo",
			url:     "https://github.com/owner/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseGitHubURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseGitHubURL() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseGitHubURL() unexpected error = %v", err)
				return
			}

			if owner != tt.wantOwner {
				t.Errorf("ParseGitHubURL() owner = %v, want %v", owner, tt.wantOwner)
			}

			if repo != tt.wantRepo {
				t.Errorf("ParseGitHubURL() repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}
