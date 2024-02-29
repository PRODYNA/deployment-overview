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
	keyVerbose              = "VERBOSE"
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
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name         string        `json:"name"`
	Error        string        `json:"error"`
	Dirty        bool          `json:"dirty"`
	Environments []Environment `json:"environments"`
}

type Environment struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
	verbose := flag.Int("verbose", LookupEnvOrInt(keyVerbose, 0), "Verbose")

	logLevel := &slog.LevelVar{}
	if verbose != nil {
		if *verbose == 0 {
			logLevel.Set(slog.LevelInfo)
		} else {
			logLevel.Set(slog.LevelDebug)
		}
	}
	flag.Parse()

	slog.InfoContext(ctx, "Configuration", "organization", config.organization, "targetRepository", config.targetRepository, "targetRepositoryFile", config.targetRepositoryFile, "repositories", config.repositories, "githubToken", "***", config.environments, config.environments, "verbose", *verbose)

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

	organization := &Organization{Name: config.organization, Repositories: []Repository{}}
	err := iterateRepositories(ctx, gh, config, organization)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to iterate repositories", "error", err)
		return
	}
	err = render(ctx, organization)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
	}
}

func iterateRepositories(ctx context.Context, gh *github.Client, config Config, organization *Organization) error {

	// split config.repositories by comma
	repos := strings.Split(config.repositories, ",")
	for _, repo := range repos {
		slog.InfoContext(ctx, "Processing repository", "organization", config.organization, "repository", repo)
		repository := Repository{Name: repo}

		repo, _, err := gh.Repositories.Get(ctx, config.organization, repo)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get repository", "organization", config.organization, "repository", repo, "error", err)
			repository.Error = err.Error()
			continue
		}

		slog.DebugContext(ctx, "Repository", "organization", config.organization, "id", repo.GetID())
		iterateEnvironments(ctx, gh, config, &repository)

		organization.Repositories = append(organization.Repositories, repository)
	}

	return nil
}

func iterateEnvironments(ctx context.Context, gh *github.Client, config Config, repository *Repository) {
	// split config.environments by comma
	envs := strings.Split(config.environments, ",")
	environments := make([]Environment, len(envs))
	for i, env := range envs {
		slog.InfoContext(ctx, "Processing environment", "organization", config.organization, "repository", repository.Name, "environment", env)
		environment := Environment{Name: env}
		environments[i] = environment

		deployments, _, err := gh.Repositories.ListDeployments(ctx, config.organization, repository.Name, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list deployments", "organization", config.organization, "repository", repository.Name, "error", err)
			environment.Version = err.Error()
			continue
		}

		for _, deployment := range deployments {
			if deployment.Environment != nil && *deployment.Environment == env {
				environment.Version = deployment.GetSHA()
				break
			}
		}
	}
	repository.Environments = environments
}

func render(ctx context.Context, organization *Organization) error {
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
