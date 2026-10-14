package main

import (
	"context"
	"github.com/google/go-github/v92/github"
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

	gh, err := github.NewClient(
		github.WithEnterpriseURLs(c.GithubAPIURL, c.GithubAPIURL),
		github.WithAuthToken(c.GithubToken),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create github client", "error", err)
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

	slog.DebugContext(ctx, "Writing json to file", "file", c.TargetJsonFile)
	err = publish.WriteFile(c.TargetJsonFile, json)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to write json to file", "error", err)
		return err
	}
	slog.InfoContext(ctx, "Wrote json to file", "file", c.TargetJsonFile)

	md, err := organization.RenderMarkdown(ctx, string(template))
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
		return err
	}

	slog.DebugContext(ctx, "Writing md to file", "file", c.TargetMdFile)
	err = publish.WriteFile(c.TargetMdFile, []byte(md))
	if err != nil {
		slog.ErrorContext(ctx, "Unable to write md to file", "error", err)
		return err
	}
	slog.InfoContext(ctx, "Wrote md to file", "file", c.TargetMdFile)

	return nil
}
