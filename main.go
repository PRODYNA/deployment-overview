package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v59/github"
	_ "github.com/google/go-github/v59/github"
	"github.com/prodyna/deployment-overview/config"
	"github.com/prodyna/deployment-overview/publish"
	"github.com/prodyna/deployment-overview/result"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	c, err := config.CreateConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create c", "error", err)
		return
	}
	err = c.Validate()
	if err != nil {
		slog.ErrorContext(ctx, "Invalid c", "error", err)
		return
	}

	// try to load the template file
	template, err := os.ReadFile(c.TemplateFile)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to read template file", "error", err, "file", c.TemplateFile)
		return
	}

	gh := github.NewClient(nil).WithAuthToken(c.GithubToken)
	if gh == nil {
		slog.ErrorContext(ctx, "Unable to create github client")
		return
	}

	organization := &result.Organization{Title: c.Title, Repositories: []result.Repository{}}
	err = organization.CreateEnvironmentDescriptions(ctx, c)
	err = organization.IterateRepositories(ctx, gh, *c)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to iterate repositories", "error", err)
		return
	}

	json, err := organization.RenderJson(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}
	fmt.Printf("%s\n", json)

	md, err := organization.RenderMarkdown(ctx, string(template))
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}

	err = publish.PublishToGitHub(ctx, c, md, gh)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to publish to github", "error", err)
	}
}
