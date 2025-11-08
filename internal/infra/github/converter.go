package github

import (
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/google/go-github/v57/github"
)

// convertToIssue converts a GitHub issue to a domain issue
func convertToIssue(ghIssue *github.Issue) *models.Issue {
	if ghIssue == nil {
		return nil
	}

	issue := &models.Issue{
		ID:       ghIssue.GetID(),
		Number:   ghIssue.GetNumber(),
		Title:    ghIssue.GetTitle(),
		Body:     ghIssue.GetBody(),
		State:    convertToIssueState(ghIssue.GetState()),
		Comments: ghIssue.GetComments(),
		Locked:   ghIssue.GetLocked(),
		URL:      ghIssue.GetURL(),
		HTMLURL:  ghIssue.GetHTMLURL(),
	}

	if ghIssue.User != nil {
		issue.Author = convertToUser(ghIssue.User)
	}

	if len(ghIssue.Assignees) > 0 {
		issue.Assignees = make([]models.User, 0, len(ghIssue.Assignees))
		for _, assignee := range ghIssue.Assignees {
			issue.Assignees = append(issue.Assignees, convertToUser(assignee))
		}
	}

	if len(ghIssue.Labels) > 0 {
		issue.Labels = make([]models.Label, 0, len(ghIssue.Labels))
		for _, label := range ghIssue.Labels {
			issue.Labels = append(issue.Labels, convertToLabel(label))
		}
	}

	if ghIssue.Milestone != nil {
		issue.Milestone = convertToMilestone(ghIssue.Milestone)
	}

	if ghIssue.CreatedAt != nil {
		issue.CreatedAt = ghIssue.CreatedAt.Time
	}

	if ghIssue.UpdatedAt != nil {
		issue.UpdatedAt = ghIssue.UpdatedAt.Time
	}

	if ghIssue.ClosedAt != nil {
		closedAt := ghIssue.ClosedAt.Time
		issue.ClosedAt = &closedAt
	}

	return issue
}

// convertToIssues converts a slice of GitHub issues to domain issues
func convertToIssues(ghIssues []*github.Issue) []*models.Issue {
	if len(ghIssues) == 0 {
		return nil
	}

	issues := make([]*models.Issue, 0, len(ghIssues))
	for _, ghIssue := range ghIssues {
		if issue := convertToIssue(ghIssue); issue != nil {
			issues = append(issues, issue)
		}
	}

	return issues
}

// convertToIssueState converts a GitHub issue state to a domain issue state
func convertToIssueState(state string) models.IssueState {
	switch state {
	case "open":
		return models.IssueStateOpen
	case "closed":
		return models.IssueStateClosed
	case "all":
		return models.IssueStateAll
	default:
		return models.IssueStateOpen
	}
}

// convertToUser converts a GitHub user to a domain user
func convertToUser(ghUser *github.User) models.User {
	if ghUser == nil {
		return models.User{}
	}

	return models.User{
		ID:        ghUser.GetID(),
		Login:     ghUser.GetLogin(),
		Name:      ghUser.GetName(),
		Email:     ghUser.GetEmail(),
		AvatarURL: ghUser.GetAvatarURL(),
	}
}

// convertToLabel converts a GitHub label to a domain label
func convertToLabel(ghLabel *github.Label) models.Label {
	if ghLabel == nil {
		return models.Label{}
	}

	return models.Label{
		ID:          ghLabel.GetID(),
		Name:        ghLabel.GetName(),
		Color:       ghLabel.GetColor(),
		Description: ghLabel.GetDescription(),
	}
}

// convertToMilestone converts a GitHub milestone to a domain milestone
func convertToMilestone(ghMilestone *github.Milestone) *models.Milestone {
	if ghMilestone == nil {
		return nil
	}

	milestone := &models.Milestone{
		ID:           ghMilestone.GetID(),
		Number:       ghMilestone.GetNumber(),
		Title:        ghMilestone.GetTitle(),
		Description:  ghMilestone.GetDescription(),
		State:        convertToMilestoneState(ghMilestone.GetState()),
		OpenIssues:   ghMilestone.GetOpenIssues(),
		ClosedIssues: ghMilestone.GetClosedIssues(),
	}

	if ghMilestone.CreatedAt != nil {
		milestone.CreatedAt = ghMilestone.CreatedAt.Time
	}

	if ghMilestone.UpdatedAt != nil {
		milestone.UpdatedAt = ghMilestone.UpdatedAt.Time
	}

	if ghMilestone.DueOn != nil {
		dueOn := ghMilestone.DueOn.Time
		milestone.DueOn = &dueOn
	}

	return milestone
}

// convertToMilestoneState converts a GitHub milestone state to a domain milestone state
func convertToMilestoneState(state string) models.MilestoneState {
	switch state {
	case "open":
		return models.MilestoneStateOpen
	case "closed":
		return models.MilestoneStateClosed
	default:
		return models.MilestoneStateOpen
	}
}

// convertFromIssueOptions converts domain issue options to GitHub list options
func convertFromIssueOptions(opts *models.IssueOptions) *github.IssueListByRepoOptions {
	if opts == nil {
		return &github.IssueListByRepoOptions{
			State: "all",
			ListOptions: github.ListOptions{
				PerPage: 30,
			},
		}
	}

	ghOpts := &github.IssueListByRepoOptions{
		State:     string(opts.State),
		Assignee:  opts.Assignee,
		Creator:   opts.Creator,
		Mentioned: opts.Mentioned,
		Labels:    opts.Labels,
		Sort:      string(opts.Sort),
		Direction: string(opts.Direction),
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.PerPage,
		},
	}

	if opts.Since != nil {
		ghOpts.Since = *opts.Since
	}

	if ghOpts.ListOptions.PerPage == 0 {
		ghOpts.ListOptions.PerPage = 30
	}

	return ghOpts
}

// convertFromCreateIssueInput converts domain create issue input to GitHub issue request
func convertFromCreateIssueInput(input *models.CreateIssueInput) *github.IssueRequest {
	if input == nil {
		return nil
	}

	req := &github.IssueRequest{
		Title: &input.Title,
	}

	if input.Body != "" {
		req.Body = &input.Body
	}

	if len(input.Assignees) > 0 {
		req.Assignees = &input.Assignees
	}

	if len(input.Labels) > 0 {
		req.Labels = &input.Labels
	}

	if input.Milestone != 0 {
		req.Milestone = &input.Milestone
	}

	return req
}

// convertFromUpdateIssueInput converts domain update issue input to GitHub issue request
func convertFromUpdateIssueInput(input *models.UpdateIssueInput) *github.IssueRequest {
	if input == nil {
		return nil
	}

	req := &github.IssueRequest{}

	if input.Title != nil {
		req.Title = input.Title
	}

	if input.Body != nil {
		req.Body = input.Body
	}

	if input.State != nil {
		state := string(*input.State)
		req.State = &state
	}

	if input.Assignees != nil {
		req.Assignees = input.Assignees
	}

	if input.Labels != nil {
		req.Labels = input.Labels
	}

	if input.Milestone != nil {
		req.Milestone = input.Milestone
	}

	return req
}
