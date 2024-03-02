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
	"time"
)

type Release struct {
	Tag       string           `json:"Tag"`
	SHA       string           `json:"SHA"`
	Title     string           `json:"Title"`
	Timestamp github.Timestamp `json:"Timestamp"`
	Link      string           `json:"Link"`
}

type Organization struct {
	Title        string       `json:"Title"`
	Repositories []Repository `json:"Repositories"`
	LastUpdated  string       `json:"LastUpdated"`
}

type Repository struct {
	Name          string        `json:"Name"`
	Error         string        `json:"Error"`
	Environments  []Environment `json:"Environments"`
	LatestRelease Release       `json:"LatestRelease"`
	Releases      []Release     `json:"Releases"`
	PullRequests  PullRequests  `json:"PullRequests"`
	Link          string        `json:"Link"`
	Commits       Commits       `json:"Commits"`
	DefaultBranch string        `json:"DefaultBranch"`
	Tags          []Tag         `json:"Tags"`
	Workflows     Workflows     `json:"Workflows"`
}

type Tag struct {
	Name string `json:"Name"`
	SHA  string `json:"SHA"`
}

type Commits struct {
	Link    string   `json:"Link"`
	Count   int      `json:"Count"`
	Commits []Commit `json:"Commits"`
}

type Author struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type Commit struct {
	Text      string           `json:"Text"`
	SHA       string           `json:"SHA"`
	Author    Author           `json:"Author"`
	Timestamp github.Timestamp `json:"Timestamp"`
	Link      string           `json:"Link"`
}

type Workflows struct {
	Link      string     `json:"Link"`
	Count     int        `json:"Count"`
	Workflows []Worfklow `json:"Workflows"`
}

type Worfklow struct {
	Name      string           `json:"Name"`
	Timestamp github.Timestamp `json:"Timestamp"`
	Link      string           `json:"Link"`
}

type Environment struct {
	Name      string `json:"Name"`
	Version   string `json:"Version"`
	IsRelease bool   `json:"IsRelease"`
	IsCurrent bool   `json:"IsCurrent"`
}

type PullRequests struct {
	Link         string        `json:"Link"`
	Count        int           `json:"Count"`
	PullRequests []PullRequest `json:"PullRequests"`
}

type PullRequest struct {
	Title string `json:"Title"`
	Link  string `json:"Link"`
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

		// get all workflows
		workflows, _, err := gh.Actions.ListRepositoryWorkflowRuns(ctx, config.Organization, rep, &github.ListWorkflowRunsOptions{
			ListOptions: github.ListOptions{PerPage: 100},
			Status:      "waiting",
		})
		if err != nil {
			slog.ErrorContext(ctx, "Unable to list workflows", "organization", config.Organization, "repository", rep, "error", err)
			repository.Error = err.Error()
		} else {
			repository.Workflows.Link = fmt.Sprintf("https://github.com/%s/%s/actions?query=is%%3Awaiting", config.Organization, rep)
			for _, workflowrun := range workflows.WorkflowRuns {
				repository.Workflows.Workflows = append(repository.Workflows.Workflows, Worfklow{
					Name:      workflowrun.GetDisplayTitle(),
					Timestamp: workflowrun.GetCreatedAt(),
					Link:      fmt.Sprintf(workflowrun.GetHTMLURL()),
				})
			}
			repository.Workflows.Count = len(repository.Workflows.Workflows)
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
			repository.Commits.Link = fmt.Sprintf("https://github.com/%s/%s/compare/%s..HEAD", config.Organization, rep, repository.LatestRelease.Tag)
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

	organization.LastUpdated = time.Now().Format(time.RFC3339)

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
			if deployment.GetEnvironment() == env {

				// get the status for this deployment
				statuses, _, err := gh.Repositories.ListDeploymentStatuses(ctx, config.Organization, repository.Name, deployment.GetID(), nil)
				if err != nil {
					slog.ErrorContext(ctx, "Unable to list deployment statuses", "organization", config.Organization, "repository", repository.Name, "error", err)
					environment.Version = err.Error()
					continue
				}
				var thisStatus *github.DeploymentStatus
				for _, status := range statuses {
					if status.GetState() == "success" {
						thisStatus = status
					}
				}
				if thisStatus == nil {
					continue
				}

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

func (organization *Organization) RenderMarkdown(ctx context.Context, templateContent string) (string, error) {
	// render the organization to markdown
	tmpl := template.Must(template.New("organization").Parse(templateContent))
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
