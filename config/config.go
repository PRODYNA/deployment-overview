package config

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

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
	Organization         string
	TargetRepository     string
	TargetRepositoryFile string
	Repositories         string
	Environments         string
	GithubToken          string
}

func CreateConfig(ctx context.Context) (*Config, error) {
	config := Config{}
	flag.StringVar(&config.Organization, "Organization", lookupEnvOrString(keyOrganization, ""), "Organization")
	flag.StringVar(&config.TargetRepository, "TargetRepository", lookupEnvOrString(keyTargetRepository, ""), "Target Repository")
	flag.StringVar(&config.TargetRepositoryFile, "TargetRepositoryFile", lookupEnvOrString(keyTargetRepositoryFile, ""), "Target Repository File")
	flag.StringVar(&config.Repositories, "Repositories", lookupEnvOrString(keyRepositories, ""), "Repositories")
	flag.StringVar(&config.GithubToken, "GithubToken", lookupEnvOrString(keyGithubToken, ""), "Github Token")
	flag.StringVar(&config.Environments, "Environments", lookupEnvOrString(keyEnvironments, ""), "Environments")
	verbose := flag.Int("verbose", lookupEnvOrInt(keyVerbose, 0), "Verbose")

	logLevel := &slog.LevelVar{}
	if verbose != nil {
		if *verbose == 0 {
			logLevel.Set(slog.LevelInfo)
		} else {
			logLevel.Set(slog.LevelDebug)
		}
	}
	flag.Parse()

	if config.Organization == "" {
		return nil, errors.New("Organization is required")
	}

	if config.GithubToken == "" {
		return nil, errors.New("Github Token is required")
	}

	if config.Repositories == "" {
		return nil, errors.New("Repositories is required")
	}

	if config.TargetRepository == "" {
		return nil, errors.New("Target Repository is required")
	}

	if config.TargetRepositoryFile == "" {
		return nil, errors.New("Target Repository File is required")
	}

	return &config, nil
}

func (c *Config) RepositoriesAsList() []string {
	return strings.Split(c.Repositories, ",")
}

func (c *Config) EnvironmentsAsList() []string {
	return strings.Split(c.Environments, ",")
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
