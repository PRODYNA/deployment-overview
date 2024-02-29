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
	err = iterateRepositories(ctx, gh, *config, organization)
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

func iterateRepositories(ctx context.Context, gh *github.Client, config config.Config, organization *result.Organization) error {

	// split config.repositories by comma
	for _, rep := range config.RepositoriesAsList() {
		slog.InfoContext(ctx, "Processing repository", "organization", config.Organization, "repository", rep)
		repository := result.Repository{Name: rep}

		repo, _, err := gh.Repositories.Get(ctx, config.Organization, rep)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get repository", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
			continue
		}

		// get the latest release
		release, _, err := gh.Repositories.GetLatestRelease(ctx, config.Organization, rep)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get latest release", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			repository.LatestRelease = result.Release{Tag: release.GetTagName(), Title: release.GetName()}
		}

		// get all releases
		releases, _, err := gh.Repositories.ListReleases(ctx, config.Organization, rep, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list releases", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			for _, release := range releases {
				repository.Releases = append(repository.Releases, result.Release{Tag: release.GetTagName(), Title: release.GetName()})
			}
		}

		// check if the repository has open pull requests
		pulls, _, err := gh.PullRequests.List(ctx, config.Organization, rep, &github.PullRequestListOptions{State: "open"})
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list open pull requests", "organization", config.Organization, "repository", repo, "error", err)
			repository.Error = err.Error()
		} else {
			repository.OpenPulls = len(pulls)
		}

		slog.DebugContext(ctx, "Repository", "organization", config.Organization, "id", repo.GetID())
		err = iterateEnvironments(ctx, gh, config, &repository)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to iterate environments", "organization", config.Organization, "repository", repo, "error", err)
			repository.Error = err.Error()
		}

		organization.Repositories = append(organization.Repositories, repository)
	}

	return nil
}

func iterateEnvironments(ctx context.Context, gh *github.Client, config config.Config, repository *result.Repository) (err error) {
	// split config.environments by comma
	for _, env := range config.EnvironmentsAsList() {
		slog.InfoContext(ctx, "Processing environment", "organization", config.Organization, "repository", repository.Name, "environment", env)
		environment := result.Environment{Name: env}

		deployments, _, err := gh.Repositories.ListDeployments(ctx, config.Organization, repository.Name, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list deployments", "organization", config.Organization, "repository", repository.Name, "error", err)
			environment.Version = err.Error()
			continue
		}

		for _, deployment := range deployments {
			if deployment.Environment != nil && *deployment.Environment == env {
				ref := deployment.GetRef()
				if ref == "master" || ref == "main" {
					environment.Version = deployment.GetSHA()[0:7]
				} else {
					environment.Version = deployment.GetRef()
				}
				repository.Environments = append(repository.Environments, environment)
				break
			}
		}
	}

	return nil
}
