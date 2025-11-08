package github

import (
	"github.com/a1yama/tig-gh/internal/domain/models"
	"github.com/google/go-github/v57/github"
)

// convertToCommit converts a GitHub repository commit to a domain commit
func convertToCommit(ghCommit *github.RepositoryCommit) *models.Commit {
	if ghCommit == nil {
		return nil
	}

	commit := &models.Commit{
		SHA:     ghCommit.GetSHA(),
		URL:     ghCommit.GetURL(),
		Message: ghCommit.GetCommit().GetMessage(),
	}

	// Author
	if ghCommit.GetCommit().GetAuthor() != nil {
		commit.Author = models.CommitAuthor{
			Name:  ghCommit.GetCommit().GetAuthor().GetName(),
			Email: ghCommit.GetCommit().GetAuthor().GetEmail(),
			Date:  ghCommit.GetCommit().GetAuthor().GetDate().Time,
		}
	}

	// Committer
	if ghCommit.GetCommit().GetCommitter() != nil {
		commit.Committer = models.CommitAuthor{
			Name:  ghCommit.GetCommit().GetCommitter().GetName(),
			Email: ghCommit.GetCommit().GetCommitter().GetEmail(),
			Date:  ghCommit.GetCommit().GetCommitter().GetDate().Time,
		}
	}

	// Parents
	if len(ghCommit.Parents) > 0 {
		commit.Parents = make([]string, 0, len(ghCommit.Parents))
		for _, parent := range ghCommit.Parents {
			commit.Parents = append(commit.Parents, parent.GetSHA())
		}
	}

	// Tree
	if ghCommit.GetCommit().GetTree() != nil {
		commit.Tree = ghCommit.GetCommit().GetTree().GetSHA()
	}

	// CreatedAt (use author date as created at)
	if ghCommit.GetCommit().GetAuthor() != nil {
		commit.CreatedAt = ghCommit.GetCommit().GetAuthor().GetDate().Time
	}

	return commit
}

// convertToCommitDetail converts a GitHub repository commit with stats to a domain commit
func convertToCommitDetail(ghCommit *github.RepositoryCommit) *models.Commit {
	commit := convertToCommit(ghCommit)
	if commit == nil {
		return nil
	}

	// Stats
	if ghCommit.Stats != nil {
		commit.Stats = &models.CommitStats{
			Additions: ghCommit.Stats.GetAdditions(),
			Deletions: ghCommit.Stats.GetDeletions(),
			Total:     ghCommit.Stats.GetTotal(),
		}
	}

	// Files
	if len(ghCommit.Files) > 0 {
		commit.Files = make([]*models.DiffFile, 0, len(ghCommit.Files))
		for _, file := range ghCommit.Files {
			commit.Files = append(commit.Files, convertToDiffFile(file))
		}
	}

	return commit
}

// convertToCommits converts a slice of GitHub repository commits to domain commits
func convertToCommits(ghCommits []*github.RepositoryCommit) []*models.Commit {
	if len(ghCommits) == 0 {
		return nil
	}

	commits := make([]*models.Commit, 0, len(ghCommits))
	for _, ghCommit := range ghCommits {
		if commit := convertToCommit(ghCommit); commit != nil {
			commits = append(commits, commit)
		}
	}

	return commits
}

// convertToDiffFile converts a GitHub commit file to a domain diff file
func convertToDiffFile(ghFile *github.CommitFile) *models.DiffFile {
	if ghFile == nil {
		return nil
	}

	return &models.DiffFile{
		Filename:  ghFile.GetFilename(),
		Status:    convertToFileStatus(ghFile.GetStatus()),
		Additions: ghFile.GetAdditions(),
		Deletions: ghFile.GetDeletions(),
		Changes:   ghFile.GetChanges(),
		Patch:     ghFile.GetPatch(),
	}
}

// convertToFileStatus converts a GitHub file status to a domain file status
func convertToFileStatus(status string) models.FileStatus {
	switch status {
	case "added":
		return models.FileStatusAdded
	case "modified":
		return models.FileStatusModified
	case "removed":
		return models.FileStatusRemoved
	case "renamed":
		return models.FileStatusRenamed
	default:
		return models.FileStatusModified
	}
}

// convertToComparison converts a GitHub commits comparison to a domain comparison
func convertToComparison(ghComparison *github.CommitsComparison) *models.Comparison {
	if ghComparison == nil {
		return nil
	}

	comparison := &models.Comparison{
		Status:       convertToComparisonStatus(ghComparison.GetStatus()),
		AheadBy:      ghComparison.GetAheadBy(),
		BehindBy:     ghComparison.GetBehindBy(),
		TotalCommits: ghComparison.GetTotalCommits(),
	}

	// Base commit
	if ghComparison.BaseCommit != nil {
		comparison.BaseCommit = convertToCommit(ghComparison.BaseCommit)
	}

	// Merge commit
	if ghComparison.MergeBaseCommit != nil {
		comparison.MergeCommit = convertToCommit(ghComparison.MergeBaseCommit)
	}

	// Commits
	if len(ghComparison.Commits) > 0 {
		comparison.Commits = convertToCommits(ghComparison.Commits)
	}

	// Files
	if len(ghComparison.Files) > 0 {
		comparison.Files = make([]*models.DiffFile, 0, len(ghComparison.Files))
		for _, file := range ghComparison.Files {
			comparison.Files = append(comparison.Files, convertToDiffFile(file))
		}
	}

	return comparison
}

// convertToComparisonStatus converts a GitHub comparison status to a domain comparison status
func convertToComparisonStatus(status string) models.ComparisonStatus {
	switch status {
	case "identical":
		return models.ComparisonStatusIdentical
	case "ahead":
		return models.ComparisonStatusAhead
	case "behind":
		return models.ComparisonStatusBehind
	case "diverged":
		return models.ComparisonStatusDiverged
	default:
		return models.ComparisonStatusDiverged
	}
}

// convertToBranch converts a GitHub branch to a domain branch
func convertToBranch(ghBranch *github.Branch) *models.Branch {
	if ghBranch == nil {
		return nil
	}

	return &models.Branch{
		Name: ghBranch.GetName(),
		SHA:  ghBranch.GetCommit().GetSHA(),
	}
}

// convertToBranches converts a slice of GitHub branches to domain branches
func convertToBranches(ghBranches []*github.Branch) []*models.Branch {
	if len(ghBranches) == 0 {
		return nil
	}

	branches := make([]*models.Branch, 0, len(ghBranches))
	for _, ghBranch := range ghBranches {
		if branch := convertToBranch(ghBranch); branch != nil {
			branches = append(branches, branch)
		}
	}

	return branches
}

// convertFromCommitOptions converts domain commit options to GitHub commit list options
func convertFromCommitOptions(opts *models.CommitOptions) *github.CommitsListOptions {
	if opts == nil {
		return &github.CommitsListOptions{
			ListOptions: github.ListOptions{
				PerPage: 30,
			},
		}
	}

	ghOpts := &github.CommitsListOptions{
		SHA:    opts.SHA,
		Path:   opts.Path,
		Author: opts.Author,
		ListOptions: github.ListOptions{
			Page:    opts.Page,
			PerPage: opts.PerPage,
		},
	}

	if opts.Since != nil {
		ghOpts.Since = *opts.Since
	}

	if opts.Until != nil {
		ghOpts.Until = *opts.Until
	}

	if ghOpts.ListOptions.PerPage == 0 {
		ghOpts.ListOptions.PerPage = 30
	}

	return ghOpts
}
