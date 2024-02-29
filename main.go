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

}
