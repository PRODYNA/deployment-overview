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
	keyEnvironmentLinks     = "ENVIRONMENT_LINKS"
	keyVerbose              = "VERBOSE"
	keyTemplateFile         = "TEMPLATE_FILE"
	keyTitle                = "TITLE"
)

type Config struct {
	Organization         string
	TargetRepository     string
	TargetRepositoryFile string
	Repositories         string
	Environments         string
	EnvironmentLinks     string
	GithubToken          string
	TemplateFile         string
	Title                string
}

func CreateConfig(ctx context.Context) (*Config, error) {
	c := Config{}
	flag.StringVar(&c.Organization, "organization", lookupEnvOrString(keyOrganization, ""), "The GitHub Organization to query for repositories.")
	flag.StringVar(&c.TargetRepository, "target-repository", lookupEnvOrString(keyTargetRepository, ""), "The target repository to commit the result to.")
	flag.StringVar(&c.TargetRepositoryFile, "target-repository-file", lookupEnvOrString(keyTargetRepositoryFile, ""), "The target repository file to commit the result to.")
	flag.StringVar(&c.Repositories, "repositories", lookupEnvOrString(keyRepositories, ""), "Repositories to query. Comma separated list.")
	flag.StringVar(&c.GithubToken, "github-token", lookupEnvOrString(keyGithubToken, ""), "The GitHub Token to use for authentication.")
	flag.StringVar(&c.Environments, "environments", lookupEnvOrString(keyEnvironments, ""), "Environments to query. Comma separated list.")
	flag.StringVar(&c.EnvironmentLinks, "environment-links", lookupEnvOrString(keyEnvironmentLinks, ""), "Links to environments. Comma separated list.")
	flag.StringVar(&c.TemplateFile, "template-file", lookupEnvOrString(keyTemplateFile, "template/default.tpl"), "The template file to use for rendering the result. Defaults to 'template/default.tpl'.")
	flag.StringVar(&c.Title, "title", lookupEnvOrString(keyTitle, "Organization Overview"), "The title to use for the result. Defaults to 'Organization Overview'.")
	verbose := flag.Int("verbose", lookupEnvOrInt(keyVerbose, 0), "Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE.")

	logLevel := &slog.LevelVar{}
	if verbose != nil {
		if *verbose == 0 {
			logLevel.Set(slog.LevelInfo)
		} else {
			logLevel.Set(slog.LevelDebug)
		}
	}
	flag.Parse()

	slog.InfoContext(ctx, "Configuration",
		"Organization", c.Organization,
		"Repositories", c.RepositoriesAsList(),
		"Environments", c.EnvironmentsAsList(),
		"EnvironmentLinks", c.EnvironmentLinksAsList(),
		"TargetRepository", c.TargetRepository,
		"TargetRepositoryFile", c.TargetRepositoryFile,
		"TemplateFile", c.TemplateFile,
		"Title", c.Title)

	return &c, nil
}

func (c *Config) Validate() error {
	if c.Organization == "" {
		return errors.New("Organization is required")
	}

	if c.GithubToken == "" {
		return errors.New("Github Token is required")
	}

	if c.Repositories == "" {
		return errors.New("Repositories is required")
	}

	if c.TargetRepository == "" {
		return errors.New("Target Repository is required")
	}

	if c.TargetRepositoryFile == "" {
		return errors.New("Target Repository File is required")
	}

	return nil

}

func (c *Config) RepositoriesAsList() []string {
	return strings.Split(c.Repositories, ",")
}

func (c *Config) EnvironmentsAsList() []string {
	return strings.Split(c.Environments, ",")
}

func (c *Config) EnvironmentLinksAsList() []string {
	return strings.Split(c.EnvironmentLinks, ",")
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
