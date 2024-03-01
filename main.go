package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v59/github"
	_ "github.com/google/go-github/v59/github"
	"github.com/prodyna/deployment-overview/config"
	"github.com/prodyna/deployment-overview/result"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	config, err := config.CreateConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create configuration", "error", err)
		return
	}
	slog.InfoContext(ctx, "Configuration",
		"Organization", config.Organization,
		"Repositories", config.RepositoriesAsList(),
		"Environments", config.EnvironmentsAsList(),
		"TargetRepository", config.TargetRepository,
		"TargetRepositoryFile", config.TargetRepositoryFile)

	gh := github.NewClient(nil).WithAuthToken(config.GithubToken)
	if gh == nil {
		slog.ErrorContext(ctx, "Unable to create github client")
		return
	}

	organization := &result.Organization{Name: config.Organization, Repositories: []result.Repository{}}
	err = organization.IterateRepositories(ctx, gh, *config)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to iterate repositories", "error", err)
		return
	}

	json, err := organization.RenderJson(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}
	fmt.Printf("%s\n", json)

	md, err := organization.RenderMarkdown(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}
	filename := "test.md"
	err = os.WriteFile(filename, []byte(md), 0644)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to write file", "error", err)
	} else {
		slog.InfoContext(ctx, "File written", "file", filename)
	}

	// commit to github using the API
	slog.InfoContext(ctx, "Committing to github", "repository", config.TargetRepository, "organization", config.Organization, "file", config.TargetRepositoryFile)
	slog.DebugContext(ctx, "Getting target repository", "repository", config.TargetRepository, "organization", config.Organization)
	repo, _, err := gh.Repositories.Get(ctx, config.Organization, config.TargetRepository)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to get target repository", "error", err, "repository", config.TargetRepository, "organization", config.Organization)
		return
	}
	slog.DebugContext(ctx, "Default branch", "branch", *repo.DefaultBranch, "repository", config.TargetRepository, "organization", config.Organization)
	defaultBranch := *repo.DefaultBranch

	slog.DebugContext(ctx, "Getting branch", "branch", defaultBranch, "repository", config.TargetRepository, "organization", config.Organization)
	branch, _, err := gh.Repositories.GetBranch(ctx, config.Organization, config.TargetRepository, defaultBranch, 0)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to get branch", "error", err, "branch", defaultBranch)
		return
	}
	slog.DebugContext(ctx, "Branch", "branch", branch, "repository", config.TargetRepository, "organization", config.Organization)

	// create blob
	slog.DebugContext(ctx, "Creating blob", "repository", config.TargetRepository, "organization", config.Organization, "file", config.TargetRepositoryFile)
	blob, _, err := gh.Git.CreateBlob(ctx, config.Organization, config.TargetRepository, &github.Blob{
		Content: &md,
		Size:    github.Int(len(md)),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create blob", "error", err)
		return
	}
	slog.DebugContext(ctx, "Blob created", "repository", config.TargetRepository, "organization", config.Organization)

	treeEntry := []*github.TreeEntry{
		{
			Path: github.String(config.TargetRepositoryFile),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  blob.SHA,
		},
	}

	// create tree
	slog.DebugContext(ctx, "Creating tree", "repository", config.TargetRepository, "organization", config.Organization)
	tree, _, err := gh.Git.CreateTree(ctx, config.Organization, config.TargetRepository, *branch.Commit.SHA, treeEntry)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create tree", "error", err, "repository", config.TargetRepository, "organization", config.Organization)
		return
	}
	slog.DebugContext(ctx, "Tree created", "repository", config.TargetRepository, "organization", config.Organization)

	// create commit
	slog.DebugContext(ctx, "Creating commit", "repository", config.TargetRepository, "organization", config.Organization)
	commit, _, err := gh.Git.CreateCommit(ctx, config.Organization, config.TargetRepository, &github.Commit{
		Message: github.String("Updated deployment overview"),
		Tree:    tree,
		Parents: []*github.Commit{
			{
				SHA: branch.Commit.SHA,
			},
		},
	}, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create commit", "error", err, "repository", config.TargetRepository, "organization", config.Organization)
		return
	}
	slog.DebugContext(ctx, "Commit created", "repository", config.TargetRepository, "organization", config.Organization)

	// update branch
	slog.DebugContext(ctx, "Updating branch", "repository", config.TargetRepository, "organization", config.Organization)
	_, _, err = gh.Git.UpdateRef(ctx, config.Organization, config.TargetRepository, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/heads/%s", defaultBranch)),
		Object: &github.GitObject{
			SHA: commit.SHA,
		},
	}, false)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to update branch", "error", err, "repository", config.TargetRepository, "organization", config.Organization)
		return
	}
	slog.DebugContext(ctx, "Branch updated", "repository", config.TargetRepository, "organization", config.Organization)
	slog.InfoContext(ctx, "Updated deployment overview", "repository", config.TargetRepository, "organization", config.Organization)
}
