package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoaderLoadsMetricsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
github:
  token: test-token
metrics:
  enabled: true
  lead_time_enabled: true
  calculation_period: 720h
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadWithPath(configPath)
	if err != nil {
		t.Fatalf("LoadWithPath returned error: %v", err)
	}

	if !cfg.Metrics.Enabled {
		t.Fatalf("expected metrics to be enabled")
	}

	if !cfg.Metrics.LeadTimeEnabled {
		t.Fatalf("expected lead time metrics to be enabled")
	}

	expectedPeriod := 720 * time.Hour
	if cfg.Metrics.CalculationPeriod != expectedPeriod {
		t.Fatalf("expected calculation period %v, got %v", expectedPeriod, cfg.Metrics.CalculationPeriod)
	}
}
