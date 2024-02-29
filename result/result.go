package result

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v59/github"
	"github.com/prodyna/deployment-overview/config"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

var markdownTemplate = `
# YASM - Yet Another Skill Management

Component Status overview

{{range .Repositories}}
## [{{.Name}}]({{.Link}})

{{if .Error}}
Error: {{.Error}}
{{else}}
- Latest Release: {{.LatestRelease.Title}}

{{if .Commits.Count}}
### [Commits on {{.DefaultBranch}} after tag {{.LatestRelease.Tag}}]({{.Commits.Link}}) ({{.Commits.Count}})
{{range .Commits.Commits}}
- [{{.Text}}]({{.Link}}) by [{{.Author.Name}}]({{.Author.Link}}) on {{.Timestamp}}
{{end}}
{{end}}

{{if .PullRequests.Count}}
### [Pull Requests]({{.PullRequests.Link}}) ({{.PullRequests.Count}})
{{range .PullRequests.PullRequests}}
- [{{.Title}}]({{.Link}})
{{end}}
{{end}}

### Environments

| Environment | {{range .Environments}} {{.Name}} | {{end}}
| --- | {{range .Environments}} --- | {{end}}
| Version | {{range .Environments}} {{.Version}} | {{end}}
| Release | {{range .Environments}} {{if .IsRelease}}:heavy_check_mark:{{else}}:x:{{end}} | {{end}}
| Current | {{range .Environments}} {{if .IsCurrent}}:heavy_check_mark:{{else}}:x:{{end}} | {{end}}

{{if .Releases}}
### Last releases
{{range .Releases }}
- [{{.Title}}]({{.Link}}) on {{.Timestamp}}
{{end}}
{{end}}
{{end}}
{{end}}
`

type Release struct {
	Tag       string           `json:"tag"`
	SHA       string           `json:"sha"`
	Title     string           `json:"title"`
	Timestamp github.Timestamp `json:"timestamp"`
	Link      string           `json:"link"`
}

type Organization struct {
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
}

type Repository struct {
	Name          string        `json:"name"`
	Error         string        `json:"error"`
	Environments  []Environment `json:"environments"`
	LatestRelease Release       `json:"latestRelease"`
	Releases      []Release     `json:"releases"`
	PullRequests  PullRequests  `json:"pullRequests"`
	Link          string        `json:"link"`
	Commits       Commits       `json:"commits"`
	DefaultBranch string        `json:"defaultBranch"`
	Tags          []Tag         `json:"tags"`
}

type Tag struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}

type Commits struct {
	Link    string   `json:"link"`
	Count   int      `json:"count"`
	Commits []Commit `json:"commits"`
}

type Author struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type Commit struct {
	Text      string           `json:"text"`
	SHA       string           `json:"sha"`
	Author    Author           `json:"author"`
	Timestamp github.Timestamp `json:"timestamp"`
	Link      string           `json:"link"`
}

type Environment struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	IsRelease bool   `json:"isRelease"`
	IsCurrent bool   `json:"isCurrent"`
}

type PullRequests struct {
	Link         string        `json:"link"`
	Count        int           `json:"count"`
	PullRequests []PullRequest `json:"pullRequests"`
}

type PullRequest struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func (organization *Organization) IterateRepositories(ctx context.Context, gh *github.Client, config config.Config) error {

	// split config.repositories by comma
	for _, rep := range config.RepositoriesAsList() {
		slog.InfoContext(ctx, "Processing repository", "organization", config.Organization, "repository", rep)
		repository := Repository{
			Name: rep,
			Link: fmt.Sprintf("https://github.com/%s/%s", config.Organization, rep),
		}

		repo, _, err := gh.Repositories.Get(ctx, config.Organization, rep)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get repository", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
			continue
		}
		repository.DefaultBranch = repo.GetDefaultBranch()

		// get the latest release
		release, _, err := gh.Repositories.GetLatestRelease(ctx, config.Organization, rep)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to get latest release", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			repository.LatestRelease = Release{
				Tag:       release.GetTagName(),
				Title:     release.GetName(),
				Link:      release.GetHTMLURL(),
				Timestamp: release.GetPublishedAt(),
			}
		}

		// get all releases
		releases, _, err := gh.Repositories.ListReleases(ctx, config.Organization, rep, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list releases", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			for i, release := range releases {
				repository.Releases = append(repository.Releases, Release{
					Tag:       release.GetTagName(),
					Title:     release.GetName(),
					Link:      release.GetHTMLURL(),
					Timestamp: *release.CreatedAt,
				})
				if i == 2 {
					break
				}
			}
		}

		// check if the repository has open pull requests
		pulls, _, err := gh.PullRequests.List(ctx, config.Organization, rep, &github.PullRequestListOptions{State: "open"})
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list open pull requests", "organization", config.Organization, "repository", repo, "error", err)
			repository.Error = err.Error()
		} else {
			repository.PullRequests.Count = len(pulls)
			repository.PullRequests.Link = fmt.Sprintf("https://github.com/%s/%s/pulls", config.Organization, rep)

			for _, pull := range pulls {
				repository.PullRequests.PullRequests = append(repository.PullRequests.PullRequests, PullRequest{
					Title: pull.GetTitle(),
					Link:  pull.GetHTMLURL(),
				})
			}
		}

		// get all commits since the latest release
		commits, _, err := gh.Repositories.ListCommits(ctx, config.Organization, rep, &github.CommitsListOptions{
			Since: repository.LatestRelease.Timestamp.Time,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list commits", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			repository.Commits.Count = len(commits)
			repository.Commits.Link = fmt.Sprintf("https://github.com/%s/%s/commits/%s", config.Organization, rep, repository.LatestRelease.Tag)
			for _, commit := range commits {
				message := commit.GetCommit().GetMessage()
				firstLine := strings.Split(message, "\n")[0]
				repository.Commits.Commits = append(repository.Commits.Commits, Commit{
					Text: firstLine,
					SHA:  commit.GetSHA(),
					Author: Author{
						Name: commit.GetAuthor().GetLogin(),
						Link: fmt.Sprintf("https://github.com/%s", commit.GetAuthor().GetLogin()),
					},
					Timestamp: commit.GetCommit().GetAuthor().GetDate(),
					Link:      fmt.Sprintf("https://github.com/%s/%s/commit/%s", config.Organization, rep, commit.GetSHA()),
				})
			}
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
					environment.IsRelease = false
				} else {
					environment.Version = deployment.GetRef()
					environment.IsRelease = true
				}
				if environment.Version == repository.LatestRelease.Tag {
					environment.IsCurrent = true
				} else {
					environment.IsCurrent = false
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

func (repository *Repository) HasTag(tagName string) bool {
	for _, tag := range repository.Tags {
		if tag.Name == tagName {
			return true
		}
	}
	return false
}
