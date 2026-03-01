package export

// テンプレートに渡すトップレベルデータ
type templateData struct {
	Overall          overallData
	PRLeadTimes      *prLeadTimesData
	ReviewPhases     *reviewPhasesData
	DayOfWeek        *dayOfWeekData
	WeeklyComparison *weeklyComparisonData
	QualityIssues    *qualityIssuesData
	StagnantPRs      *stagnantPRsData
	RepositoryStats  *repositoryStatsData
}

type overallData struct {
	Average string
	Median  string
	Count   int
}

type prLeadTimeEntry struct {
	Index      int
	Repository string
	Number     int
	Title      string
	LeadTime   string
	MergedAt   string
	HTMLURL    string
}

type prLeadTimesData struct {
	Entries []prLeadTimeEntry
}

type phaseEntry struct {
	Label        string
	Duration     string
	SampleCount  int
	IsBottleneck bool
}

type reviewPhasesData struct {
	Phases      []phaseEntry
	TotalLabel  string
	SampleCount int
}

type dayOfWeekEntry struct {
	Day         string
	MergeCount  int
	ReviewCount int
}

type dayOfWeekData struct {
	Days []dayOfWeekEntry
}

type weeklyComparisonData struct {
	ThisWeekReviews int
	ThisWeekMerges  int
	LastWeekReviews int
	LastWeekMerges  int
	ReviewChange    string
	MergeChange     string
}

type qualityIssueEntry struct {
	Repository string
	Number     int
	Title      string
	IssueType  string
	Details    string
}

type qualityIssuesData struct {
	HighPriority   []qualityIssueEntry
	MediumPriority []qualityIssueEntry
}

type stagnantPREntry struct {
	Index      int
	Repository string
	Number     int
	Title      string
	Age        string
}

type stagnantPRsData struct {
	TotalStagnant int
	Threshold     string
	Entries       []stagnantPREntry
}

type repoStatEntry struct {
	Name    string
	Average string
	Median  string
	Count   int
}

type repositoryStatsData struct {
	Repos []repoStatEntry
}
