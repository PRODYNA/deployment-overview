package config

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	keyOrganization     = "ORGANIZATION"
	keyRepositories     = "REPOSITORIES"
	keyGithubToken      = "GITHUB_TOKEN"
	keyGithubAPIURL     = "GITHUB_API_URL"
	keyGithubServerURL  = "GITHUB_SERVER_URL"
	keyEnvironments     = "ENVIRONMENTS"
	keyEnvironmentLinks = "ENVIRONMENT_LINKS"
	keyVerbose          = "VERBOSE"
	keyTemplateFile     = "TEMPLATE_FILE"
	keyTitle            = "TITLE"
	keyTargetJsonFile   = "TARGET_JSON_FILE"
	keyTargetMdFile     = "TARGET_MD_FILE"
)

type Config struct {
	Organization     string
	TargetJsonFile   string
	TargetMdFile     string
	Repositories     string
	Environments     string
	EnvironmentLinks string
	GithubToken      string
	GithubAPIURL     string
	GithubServerURL  string
	TemplateFile     string
	Title            string
}

func CreateConfig(ctx context.Context) (*Config, error) {
	c := Config{}
	flag.StringVar(&c.Organization, "organization", lookupEnvOrString(keyOrganization, ""), "The GitHub Organization to query for repositories.")
	flag.StringVar(&c.TargetJsonFile, "target-json-file", lookupEnvOrString(keyTargetJsonFile, "deployment-overview.json"), "The target json file to commit the result to.")
	flag.StringVar(&c.TargetMdFile, "target-md-file", lookupEnvOrString(keyTargetMdFile, "deployment-overview.md"), "The target md file to commit the result to.")
	flag.StringVar(&c.Repositories, "repositories", lookupEnvOrString(keyRepositories, ""), "Repositories to query. Comma separated list.")
	flag.StringVar(&c.GithubToken, "github-token", lookupEnvOrString(keyGithubToken, ""), "The GitHub Token to use for authentication.")
	flag.StringVar(&c.GithubAPIURL, "github-api-url", lookupEnvOrString(keyGithubAPIURL, "https://api.github.com/"), "The GitHub API URL.")
	flag.StringVar(&c.GithubServerURL, "github-server-url", lookupEnvOrString(keyGithubServerURL, "https://github.com"), "The GitHub web URL used for links.")
	flag.StringVar(&c.Environments, "environments", lookupEnvOrString(keyEnvironments, ""), "Environments to query. Comma separated list.")
	flag.StringVar(&c.EnvironmentLinks, "environment-links", lookupEnvOrString(keyEnvironmentLinks, ""), "Links to environments. Comma separated list.")
	flag.StringVar(&c.TemplateFile, "template-file", lookupEnvOrString(keyTemplateFile, "template/default.tpl"), "The template file to use for rendering the result. Defaults to 'template/default.tpl'.")
	flag.StringVar(&c.Title, "title", lookupEnvOrString(keyTitle, "Organization Overview"), "The title to use for the result. Defaults to 'Organization Overview'.")
	verbose := flag.Int("verbose", lookupEnvOrInt(keyVerbose, 0), "Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE.")

	flag.Parse()
	level := slog.LevelInfo
	if *verbose == 1 {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(handler)

	slog.InfoContext(ctx, "Configuration",
		"Organization", c.Organization,
		"Repositories", c.RepositoriesAsList(),
		"GitHubAPIURL", c.GithubAPIURL,
		"GitHubServerURL", c.GithubServerURL,
		"Environments", c.EnvironmentsAsList(),
		"EnvironmentLinks", c.EnvironmentLinksAsList(),
		"TemplateFile", c.TemplateFile,
		"TargetJsonFile", c.TargetJsonFile,
		"TargetMdFile", c.TargetMdFile,
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

	if err := validateURL(c.GithubAPIURL); err != nil {
		return errors.New("GitHub API URL must be an absolute URL with a host")
	}

	if err := validateURL(c.GithubServerURL); err != nil {
		return errors.New("GitHub server URL must be an absolute URL with a host")
	}

	return nil

}

func validateURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("invalid URL")
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
