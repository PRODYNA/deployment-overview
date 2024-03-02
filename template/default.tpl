# {{.Title}}

Component Status overview. Last update {{.LastUpdated}}

| Repository | Commits | PRs | Latest Release | {{ range .EnvironmentDescriptions }} [{{.Name}}]({{.Link}}) | {{end}}
| --- | --- | --- | -- | {{ range .EnvironmentDescriptions }} --- | {{end}}
{{range .Repositories}}| [{{.Name}}]({{.Link}}) | {{if .Commits.Count}}:red_square:{{else}}:green_square:{{end}} [{{.Commits.Count}}]({{.Commits.Link}}) | {{if .PullRequests.Count}}:yellow_square:{{else}}:green_square:{{end}} [{{.PullRequests.Count}}]({{.PullRequests.Link}}) | {{.LatestRelease.Tag}} | {{range .Environments}} {{if .IsCurrent}}:green_square:{{else}}:red_square:{{end}} {{.Version}} | {{end}}
{{end}}

{{range .Repositories}}
## [{{.Name}}]({{.Link}}) {{.LatestRelease.Tag}}

{{if .Error}}
> [!WARNING]
> {{.Error}}
{{else}}

{{if .Commits.Count}}
### [Commits on {{.DefaultBranch}} since {{.LatestRelease.Tag}}]({{.Commits.Link}}) ({{.Commits.Count}})
{{range .Commits.Commits}}
- [{{.Text}}]({{.Link}}) by [{{.Author.Name}}]({{.Author.Link}}) on {{.Timestamp}}
{{end}}
{{end}}

{{if .PullRequests.Count}}
### [Open Pull Requests]({{.PullRequests.Link}}) ({{.PullRequests.Count}})
{{range .PullRequests.PullRequests}}
- [{{.Title}}]({{.Link}})
{{end}}
{{end}}

### Environments

| Environment | {{range .Environments}} {{.Name}} | {{end}}
| --- | {{range .Environments}} --- | {{end}}
| Version | {{range .Environments}} {{.Version}} | {{end}}
| Release | {{range .Environments}} {{if .IsRelease}}:green_square:{{else}}:red_square:{{end}} | {{end}}
| Current | {{range .Environments}} {{if .IsCurrent}}:green_square:{{else}}:red_square:{{end}} | {{end}}

{{if .Releases}}
### Last releases
{{range .Releases }}
- [{{.Title}}]({{.Link}}) on {{.Timestamp}}
{{end}}
{{end}}

{{if .Workflows.Count}}
### [Workflows requiring approval]({{.Workflows.Link}}) ({{.Workflows.Count}})
{{range .Workflows.Workflows}}
- [{{.Name}}]({{.Link}}) created on {{.Timestamp}}
{{end}}
{{end}}
{{end}}
{{end}}

---

Created with :heart: by the GitHub Action [deployment-overview](https://github.com/prodyna/deployment-overview) by [@dkrizic](https://github.com/dkrizic)

`
