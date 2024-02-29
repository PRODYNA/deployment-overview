package result

import (
	"bytes"
	"context"
	"encoding/json"
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
Latest Tag: {{.LatestTag}}
Commits After Tag: {{.CommitsAfterTag}}
Latest Release: {{.LatestRelease.Tag}} - {{.LatestRelease.Title}}
Open Pulls: {{.OpenPulls}}

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
