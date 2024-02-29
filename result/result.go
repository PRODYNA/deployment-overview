package result

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/go-github/v59/github"
	"github.com/prodyna/deployment-overview/config"
	"log/slog"
	"os"
	"text/template"
)

var markdownTemplate = `
# {{.Name}}

{{range .Repositories}}
## {{.Name}}

{{if .Error}}
Error: {{.Error}}
{{else}}
- Latest Tag: {{.LatestTag}}
- Commits After Tag: {{.CommitsAfterTag}}
- Latest Release: {{.LatestRelease.Tag}} - {{.LatestRelease.Title}}
- Open Pulls: {{.OpenPulls}}

### Environments
{{range .Environments}}
- {{.Name}}: {{.Version}}
{{end}}
{{end}}
{{end}}
`

type Release struct {
	Tag   string `json:"tag"`
	Title string `json:"title"`
}

type Organization struct {
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name            string        `json:"name"`
	Error           string        `json:"error"`
	Environments    []Environment `json:"environments"`
	OpenPulls       int           `json:"openPulls"`
	LatestTag       string        `json:"latestTag"`
	CommitsAfterTag int           `json:"commitsAfterTag"`
	LatestRelease   Release       `json:"latestRelease"`
	Releases        []Release     `json:"releases"`
}

type Environment struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (organization *Organization) IterateRepositories(ctx context.Context, gh *github.Client, config config.Config) error {

	// split config.repositories by comma
	for _, rep := range config.RepositoriesAsList() {
		slog.InfoContext(ctx, "Processing repository", "organization", config.Organization, "repository", rep)
		repository := Repository{Name: rep}

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
			repository.LatestRelease = Release{Tag: release.GetTagName(), Title: release.GetName()}
		}

		// get all releases
		releases, _, err := gh.Repositories.ListReleases(ctx, config.Organization, rep, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list releases", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			for _, release := range releases {
				repository.Releases = append(repository.Releases, Release{Tag: release.GetTagName(), Title: release.GetName()})
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
		err = repository.IterateEnvironments(ctx, gh, config)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to iterate environments", "organization", config.Organization, "repository", repo, "error", err)
			repository.Error = err.Error()
		}

		organization.Repositories = append(organization.Repositories, repository)
	}

	return nil
}

func (repository *Repository) IterateEnvironments(ctx context.Context, gh *github.Client, config config.Config) (err error) {
	// split config.environments by comma
	for _, env := range config.EnvironmentsAsList() {
		slog.InfoContext(ctx, "Processing environment", "organization", config.Organization, "repository", repository.Name, "environment", env)
		environment := Environment{Name: env}

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

func (organization *Organization) RenderJson(ctx context.Context) (result []byte, err error) {
	output, err := json.MarshalIndent(organization, "", "  ")
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
		return nil, err
	}
	os.Stdout.Write(output)
	return output, nil
}

func (organization *Organization) RenderMarkdown(ctx context.Context) (string, error) {
	// render the organization to markdown
	tmpl := template.Must(template.New("organization").Parse(markdownTemplate))
	// execute template to a string
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, organization)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to render organization", "error", err)
		return "", err
	}
	return buffer.String(), nil
}
