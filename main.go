package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-github/v59/github"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)
import _ "github.com/google/go-github/v59/github"

const (
	keyOrganization         = "ORGANIZATION"
	keyTargetRepository     = "TARGET_REPOSITORY"
	keyTargetRepositoryFile = "TARGET_REPOSITORY_FILE"
	keyRepositories         = "REPOSITORIES"
	keyGithubToken          = "GITHUB_TOKEN"
	keyEnvironments         = "ENVIRONMENTS"
)

type Config struct {
	organization         string
	targetRepository     string
	targetRepositoryFile string
	repositories         string
	environments         string
	githubToken          string
}

type Organization struct {
	Name         string
	Repositories []Repository
}

type Repository struct {
	Name         string
	Error        string
	Dirty        bool
	Environments []Environment
}

type Environment struct {
	Name    string
	Version string
}

func main() {
	ctx := context.Background()

	config := Config{}
	flag.StringVar(&config.organization, "organization", LookupEnvOrString(keyOrganization, ""), "Organization")
	flag.StringVar(&config.targetRepository, "targetRepository", LookupEnvOrString(keyTargetRepository, ""), "Target Repository")
	flag.StringVar(&config.targetRepositoryFile, "targetRepositoryFile", LookupEnvOrString(keyTargetRepositoryFile, ""), "Target Repository File")
	flag.StringVar(&config.repositories, "repositories", LookupEnvOrString(keyRepositories, ""), "Repositories")
	flag.StringVar(&config.githubToken, "githubToken", LookupEnvOrString(keyGithubToken, ""), "Github Token")
	flag.StringVar(&config.environments, "environments", LookupEnvOrString(keyEnvironments, ""), "Environments")
	flag.Parse()

	slog.InfoContext(ctx, "Configuration", "organization", config.organization, "targetRepository", config.targetRepository, "targetRepositoryFile", config.targetRepositoryFile, "repositories", config.repositories, "githubToken", "***", config.environments, config.environments)

	if config.organization == "" {
		slog.ErrorContext(ctx, "Organization is required")
		return
	}

	if config.githubToken == "" {
		slog.ErrorContext(ctx, "Github Token is required")
		return
	}

	if config.repositories == "" {
		slog.ErrorContext(ctx, "Repositories is required")
		return
	}

	if config.targetRepository == "" {
		slog.ErrorContext(ctx, "Target Repository is required")
		return
	}

	if config.targetRepositoryFile == "" {
		slog.ErrorContext(ctx, "Target Repository File is required")
		return
	}

	gh := github.NewClient(nil).WithAuthToken(config.githubToken)
	if gh == nil {
		slog.ErrorContext(ctx, "Unable to create github client")
		return
	}

	organization := Organization{Name: config.organization}

	// split config.repositories by comma
	repos := strings.Split(config.repositories, ",")
	for _, repo := range repos {
		slog.InfoContext(ctx, "Processing repository", "organization", config.organization, "repository", repo)
		repository := Repository{Name: repo}
		organization.Repositories = append(organization.Repositories, repository)

		repo, _, err := gh.Repositories.Get(ctx, config.organization, repo)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get repository", "organization", config.organization, "repository", repo, "error", err)
			repository.Error = err.Error()
			continue
		}

		slog.DebugContext(ctx, "Repository", "organization", config.organization, "id", repo.GetID())
	}

	err := render(ctx, organization)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}
}

func render(ctx context.Context, organization Organization) error {
	// create json and print it
	output, err := json.MarshalIndent(organization, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
