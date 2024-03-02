package publish

import (
	"context"
	"fmt"
	"github.com/google/go-github/v59/github"
	"github.com/prodyna/deployment-overview/config"
	"log/slog"
)

func PublishToGitHub(ctx context.Context, c *config.Config, md string, gh *github.Client) (err error) {
	// commit to github using the API
	slog.InfoContext(ctx, "Committing to github", "repository", c.TargetRepository, "organization", c.Organization, "file", c.TargetRepositoryFile)
	slog.DebugContext(ctx, "Getting target repository", "repository", c.TargetRepository, "organization", c.Organization)
	repo, _, err := gh.Repositories.Get(ctx, c.Organization, c.TargetRepository)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to get target repository", "error", err, "repository", c.TargetRepository, "organization", c.Organization)
		return err
	}
	slog.DebugContext(ctx, "Default branch", "branch", *repo.DefaultBranch, "repository", c.TargetRepository, "organization", c.Organization)
	defaultBranch := *repo.DefaultBranch

	slog.DebugContext(ctx, "Getting branch", "branch", defaultBranch, "repository", c.TargetRepository, "organization", c.Organization)
	branch, _, err := gh.Repositories.GetBranch(ctx, c.Organization, c.TargetRepository, defaultBranch, 0)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to get branch", "error", err, "branch", defaultBranch)
		return err
	}
	slog.DebugContext(ctx, "Branch", "branch", branch, "repository", c.TargetRepository, "organization", c.Organization)

	// create blob
	slog.DebugContext(ctx, "Creating blob", "repository", c.TargetRepository, "organization", c.Organization, "file", c.TargetRepositoryFile)
	blob, _, err := gh.Git.CreateBlob(ctx, c.Organization, c.TargetRepository, &github.Blob{
		Content: &md,
		Size:    github.Int(len(md)),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create blob", "error", err)
		return err
	}
	slog.DebugContext(ctx, "Blob created", "repository", c.TargetRepository, "organization", c.Organization)

	treeEntry := []*github.TreeEntry{
		{
			Path: github.String(c.TargetRepositoryFile),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  blob.SHA,
		},
	}

	// create tree
	slog.DebugContext(ctx, "Creating tree", "repository", c.TargetRepository, "organization", c.Organization)
	tree, _, err := gh.Git.CreateTree(ctx, c.Organization, c.TargetRepository, *branch.Commit.SHA, treeEntry)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create tree", "error", err, "repository", c.TargetRepository, "organization", c.Organization)
		return err
	}
	slog.DebugContext(ctx, "Tree created", "repository", c.TargetRepository, "organization", c.Organization)

	// create commit
	slog.DebugContext(ctx, "Creating commit", "repository", c.TargetRepository, "organization", c.Organization)
	commit, _, err := gh.Git.CreateCommit(ctx, c.Organization, c.TargetRepository, &github.Commit{
		Message: github.String("Updated deployment overview"),
		Tree:    tree,
		Parents: []*github.Commit{
			{
				SHA: branch.Commit.SHA,
			},
		},
	}, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create commit", "error", err, "repository", c.TargetRepository, "organization", c.Organization)
		return err
	}
	slog.DebugContext(ctx, "Commit created", "repository", c.TargetRepository, "organization", c.Organization)

	// update branch
	slog.DebugContext(ctx, "Updating branch", "repository", c.TargetRepository, "organization", c.Organization)
	_, _, err = gh.Git.UpdateRef(ctx, c.Organization, c.TargetRepository, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", defaultBranch)),
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}, false)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to update branch", "error", err, "repository", c.TargetRepository, "organization", c.Organization)
		return err
	}
	slog.DebugContext(ctx, "Branch updated", "repository", c.TargetRepository, "organization", c.Organization)
	slog.InfoContext(ctx, "Updated deployment overview", "repository", c.TargetRepository, "organization", c.Organization)

	return nil
}
