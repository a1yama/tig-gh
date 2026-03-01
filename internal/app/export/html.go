package export

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/a1yama/tig-gh/internal/domain/models"
)

// GenerateHTML はメトリクスデータをスタンドアロンHTMLとして生成する。
func GenerateHTML(metrics *models.LeadTimeMetrics, cfg *models.MetricsConfig) (string, error) {
	if metrics == nil {
		return "", errors.New("metrics must not be nil")
	}
	if cfg == nil {
		return "", errors.New("config must not be nil")
	}

	data := buildTemplateData(metrics, cfg)

	tmpl, err := template.New("metrics").Parse(templateHTML)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func buildTemplateData(metrics *models.LeadTimeMetrics, cfg *models.MetricsConfig) templateData {
	data := templateData{
		Overall: buildOverall(metrics),
	}

	if cfg.ShowPRLeadTimes {
		data.PRLeadTimes = buildPRLeadTimes(metrics)
	}
	if cfg.ShowReviewPhases {
		data.ReviewPhases = buildReviewPhases(metrics)
	}
	if cfg.ShowDayOfWeek {
		data.DayOfWeek = buildDayOfWeek(metrics, cfg)
	}
	if cfg.ShowWeeklyComparison {
		data.WeeklyComparison = buildWeeklyComparison(metrics)
	}
	if cfg.ShowQualityIssues {
		data.QualityIssues = buildQualityIssues(metrics)
	}
	if cfg.ShowStagnantPRs {
		data.StagnantPRs = buildStagnantPRs(metrics)
	}
	if cfg.ShowRepositoryStats {
		data.RepositoryStats = buildRepositoryStats(metrics)
	}

	return data
}
