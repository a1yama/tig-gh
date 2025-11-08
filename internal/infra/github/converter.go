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

	// GitHub APIの Issues.ListByRepo は Pull Request も返すため除外
	// Pull Request には PullRequestLinks フィールドがある
	if ghIssue.PullRequestLinks != nil {
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

// convertToPullRequest converts a GitHub pull request to a domain pull request
func convertToPullRequest(ghPR *github.PullRequest) *models.PullRequest {
	if ghPR == nil {
		return nil
	}

	pr := &models.PullRequest{
		ID:             ghPR.GetID(),
		Number:         ghPR.GetNumber(),
		Title:          ghPR.GetTitle(),
		Body:           ghPR.GetBody(),
		State:          convertToPRState(ghPR.GetState()),
		Mergeable:      ghPR.GetMergeable(),
		MergeableState: ghPR.GetMergeableState(),
		Merged:         ghPR.GetMerged(),
		Draft:          ghPR.GetDraft(),
		Locked:         ghPR.GetLocked(),
		Comments:       ghPR.GetComments(),
		Commits:        ghPR.GetCommits(),
		Additions:      ghPR.GetAdditions(),
		Deletions:      ghPR.GetDeletions(),
		ChangedFiles:   ghPR.GetChangedFiles(),
	}

	if ghPR.User != nil {
		pr.Author = convertToUser(ghPR.User)
	}

	if ghPR.Head != nil {
		pr.Head = models.Branch{
			Name: ghPR.Head.GetRef(),
			SHA:  ghPR.Head.GetSHA(),
		}
	}

	if ghPR.Base != nil {
		pr.Base = models.Branch{
			Name: ghPR.Base.GetRef(),
			SHA:  ghPR.Base.GetSHA(),
		}
	}

	if len(ghPR.RequestedReviewers) > 0 {
		pr.RequestedReviewers = make([]models.User, 0, len(ghPR.RequestedReviewers))
		for _, reviewer := range ghPR.RequestedReviewers {
			pr.RequestedReviewers = append(pr.RequestedReviewers, convertToUser(reviewer))
		}
	}

	if len(ghPR.Assignees) > 0 {
		pr.Assignees = make([]models.User, 0, len(ghPR.Assignees))
		for _, assignee := range ghPR.Assignees {
			pr.Assignees = append(pr.Assignees, convertToUser(assignee))
		}
	}

	if len(ghPR.Labels) > 0 {
		pr.Labels = make([]models.Label, 0, len(ghPR.Labels))
		for _, label := range ghPR.Labels {
			pr.Labels = append(pr.Labels, convertToLabel(label))
		}
	}

	if ghPR.Milestone != nil {
		pr.Milestone = convertToMilestone(ghPR.Milestone)
	}

	if ghPR.MergedAt != nil {
		mergedAt := ghPR.MergedAt.Time
		pr.MergedAt = &mergedAt
	}

	if ghPR.MergedBy != nil {
		mergedBy := convertToUser(ghPR.MergedBy)
		pr.MergedBy = &mergedBy
	}

	if ghPR.CreatedAt != nil {
		pr.CreatedAt = ghPR.CreatedAt.Time
	}

	if ghPR.UpdatedAt != nil {
		pr.UpdatedAt = ghPR.UpdatedAt.Time
	}

	if ghPR.ClosedAt != nil {
		closedAt := ghPR.ClosedAt.Time
		pr.ClosedAt = &closedAt
	}

	return pr
}

// convertToPullRequests converts a slice of GitHub pull requests to domain pull requests
func convertToPullRequests(ghPRs []*github.PullRequest) []*models.PullRequest {
	if len(ghPRs) == 0 {
		return nil
	}

	prs := make([]*models.PullRequest, 0, len(ghPRs))
	for _, ghPR := range ghPRs {
		if pr := convertToPullRequest(ghPR); pr != nil {
			prs = append(prs, pr)
		}
	}

	return prs
}

// convertToPRState converts a GitHub PR state to a domain PR state
func convertToPRState(state string) models.PRState {
	switch state {
	case "open":
		return models.PRStateOpen
	case "closed":
		return models.PRStateClosed
	case "all":
		return models.PRStateAll
	default:
		return models.PRStateOpen
	}
}

// convertToReview converts a GitHub review to a domain review
func convertToReview(ghReview *github.PullRequestReview) *models.Review {
	if ghReview == nil {
		return nil
	}

	review := &models.Review{
		ID:    ghReview.GetID(),
		Body:  ghReview.GetBody(),
		State: convertToReviewState(ghReview.GetState()),
	}

	if ghReview.User != nil {
		review.User = convertToUser(ghReview.User)
	}

	if ghReview.SubmittedAt != nil {
		review.SubmittedAt = ghReview.SubmittedAt.Time
	}

	return review
}

// convertToReviews converts a slice of GitHub reviews to domain reviews
func convertToReviews(ghReviews []*github.PullRequestReview) []*models.Review {
	if len(ghReviews) == 0 {
		return nil
	}

	reviews := make([]*models.Review, 0, len(ghReviews))
	for _, ghReview := range ghReviews {
		if review := convertToReview(ghReview); review != nil {
			reviews = append(reviews, review)
		}
	}

	return reviews
}

// convertToReviewState converts a GitHub review state to a domain review state
func convertToReviewState(state string) models.ReviewState {
	switch state {
	case "APPROVED":
		return models.ReviewStateApproved
	case "CHANGES_REQUESTED":
		return models.ReviewStateChangesRequested
	case "COMMENTED":
		return models.ReviewStateCommented
	case "DISMISSED":
		return models.ReviewStateDismissed
	case "PENDING":
		return models.ReviewStatePending
	default:
		return models.ReviewStateCommented
	}
}

// convertFromPROptions converts domain PR options to GitHub pull request list options
func convertFromPROptions(opts *models.PROptions) *github.PullRequestListOptions {
	if opts == nil {
		return &github.PullRequestListOptions{
			State: "all",
			ListOptions: github.ListOptions{
				PerPage: 30,
			},
		}
	}

	ghOpts := &github.PullRequestListOptions{
		State:     string(opts.State),
		Head:      opts.Head,
		Base:      opts.Base,
		Sort:      string(opts.Sort),
		Direction: string(opts.Direction),
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.PerPage,
		},
	}

	if ghOpts.ListOptions.PerPage == 0 {
		ghOpts.ListOptions.PerPage = 30
	}

	return ghOpts
}

// convertFromCreatePRInput converts domain create PR input to GitHub new pull request
func convertFromCreatePRInput(input *models.CreatePRInput) *github.NewPullRequest {
	if input == nil {
		return nil
	}

	req := &github.NewPullRequest{
		Title: &input.Title,
		Head:  &input.Head,
		Base:  &input.Base,
		Draft: &input.Draft,
	}

	if input.Body != "" {
		req.Body = &input.Body
	}

	return req
}

// convertFromUpdatePRInput converts domain update PR input to GitHub pull request
func convertFromUpdatePRInput(input *models.UpdatePRInput) *github.PullRequest {
	if input == nil {
		return nil
	}

	req := &github.PullRequest{}

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

	if input.Base != nil {
		req.Base = &github.PullRequestBranch{
			Ref: input.Base,
		}
	}

	return req
}

// convertFromMergeOptions converts domain merge options to GitHub pull request merge options
func convertFromMergeOptions(opts *models.MergeOptions) *github.PullRequestOptions {
	if opts == nil {
		return &github.PullRequestOptions{
			MergeMethod: "merge",
		}
	}

	mergeMethod := string(opts.MergeMethod)
	if mergeMethod == "" {
		mergeMethod = "merge"
	}

	ghOpts := &github.PullRequestOptions{
		MergeMethod: mergeMethod,
		SHA:         opts.SHA,
	}

	return ghOpts
}
