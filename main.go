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
	err := do()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	ctx := context.Background()
	c, err := config.CreateConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create c", "error", err)
		return err
	}
	err = c.Validate()
	if err != nil {
		slog.ErrorContext(ctx, "Invalid c", "error", err)
		return err
	}

	// try to load the template file
	template, err := os.ReadFile(c.TemplateFile)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to read template file", "error", err, "file", c.TemplateFile)
		return err
	}

	gh := github.NewClient(nil).WithAuthToken(c.GithubToken)
	if gh == nil {
		slog.ErrorContext(ctx, "Unable to create github client")
		return err
	}

	organization := &result.Organization{Title: c.Title, Repositories: []result.Repository{}}
	err = organization.CreateEnvironmentDescriptions(ctx, c)
	err = organization.IterateRepositories(ctx, gh, *c)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to iterate repositories", "error", err)
		return err
	}

	json, err := organization.RenderJson(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
		return err
	}
	fmt.Printf("%s\n", json)

	md, err := organization.RenderMarkdown(ctx, string(template))
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
		return err
	}

	err = publish.PublishToGitHub(ctx, c, md, gh)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to publish to github", "error", err)
		return err
	}

	return nil
}
