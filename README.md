# GitHub Action that creates a deployment overview page

## Sample output

| Repository | Commits | PRs | Latest Release | [dev](https://dev-yasm.prodyna.com) | [staging](https://dev-yasm.prodyna.com) | [prod](https://yasm.prodyna.com) |
| --- | --- | --- | -- | --- | --- | --- |
| [yasm-backend](https://github.com/prodyna-yasm/yasm-backend) | :red_square: 1 | :yellow_square: 1 | 1.13.1 |  :red_square: e6e2948 |  :green_square: 1.13.1 |  :green_square: 1.13.1 | 
| [yasm-frontend](https://github.com/prodyna-yasm/yasm-frontend) | :red_square: 11 | :yellow_square: 2 | 1.13.5 |  :red_square: 4aa310d |  :green_square: 1.13.5 |  :green_square: 1.13.5 | 
| [yasmctl](https://github.com/prodyna-yasm/yasmctl) | :green_square: 0 | :green_square: 0 | 1.13.3 |  :green_square: 1.13.3 |  :green_square: 1.13.3 |  :green_square: 1.13.3 | 
| [yasm-proxy-odbc](https://github.com/prodyna-yasm/yasm-proxy-odbc) | :green_square: 0 | :green_square: 0 | 1.10.0 |  :green_square: 1.10.0 |  :green_square: 1.10.0 |  :green_square: 1.10.0 | 
| [yasm-integration](https://github.com/prodyna-yasm/yasm-integration) | :green_square: 0 | :yellow_square: 1 | 1.13.5 |  :green_square: 1.13.5 |  :green_square: 1.13.5 |  :green_square: 1.13.5 | 
| [yasm-geocoding](https://github.com/prodyna-yasm/yasm-geocoding) | :red_square: 2 | :green_square: 0 | 1.3.0 |  :red_square: ec76731 |  :green_square: 1.3.0 |  :green_square: 1.3.0 | 
| [yasm-data](https://github.com/prodyna-yasm/yasm-data) | :green_square: 0 | :green_square: 0 | 1.9.0 |  :green_square: 1.9.0 |  :green_square: 1.9.0 |  :green_square: 1.9.0 | 
| [yasm-gotenberg](https://github.com/prodyna-yasm/yasm-gotenberg) | :green_square: 0 | :green_square: 0 | 8.2.0-3 |  :green_square: 8.2.0-3 |  :green_square: 8.2.0-3 |  :green_square: 8.2.0-3 | 

See [EXAMPLE.md](EXAMPLE.md) for a full example.

## Requirements:

* A GitHub repository or a GitHub organization
* A GitHub Personal Access Token (PAT)
* Go 1.22 or later

### Compile the code

```shell
go buid main.go -o deployment-overview
```

### Run the code

```shell
./deployment-overview --github-token <PAT> --organization <organization> --repositories <repository1>,<repository2> --target-repository <target-repository> --target-repository-file <target-repository-file> 
```

## Parrameters

| Paramter | Environment | Required | Default                  | Example                                                                     | Description                                                                   |
| --- | --- |-------|--------------------------|-----------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| --github-token | GITHUB_TOKEN | true  | -                        | -                                                                           | The GitHub Personal Access Token (PAT)                                        |
| --environments | ENVIRONMENTS | true  | -                        | dev,staging,prod                                                            | Environments to query. Comma separated list.                                  |
| --organization | ORGANIZATION | true  | -                        | myorga                                                                      | The GitHub Organization to query for repositories.                            |
| --repositories | REPOSITORIES | true  | -                        | frontend,backend                                                            | Repositories to query. Comma separated list.                                  |
| --verbose | VERBOSE | false  | 1                        | 0                                                                           | Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE. |
| --environment-links | ENVIRONMENT_LINKS | false  | -                        | https://dev.example.com,https://staging.example.com,https://www.example.com | Links to the environments. Comma separated list.                             |
| --template-file | TEMPLATE_FILE | false  | -                        | template/default.md                                              | The template file to use.                                                    |
| --target-json-file | TARGET_JSON_FILE | false  | deployment-overview.json | -                                                                           | The target file to write the result to as JSON.                               |
| --target-md-file | TARGET_MD_FILE | false  | deployment-overview.md   | -                                                                           | The target file to write the result to as Markdown.                           |

## Use as GitHub Action

Example on using this action in a workflow:

```yaml
name: Create Overview

on:
  workflow_dispatch:
  # Every day at 07:00
  schedule:
    - cron: '0 7 * * *'

jobs:
  create-overview:
    runs-on: ubuntu-latest
    steps:
      # Checkout the existing content of thre repository
      - name: Checkout
        uses: actions/checkout@v2

      # Create directory profile if it does not exist
      - name: Create profile directory
        run: mkdir -p profile

      # Run the deployment overview action
      - name: Deployment overview
        uses: prodyna/deployment-overview@v0.3
        with:
          organization: prodyna-yasm
          repositories: yasm-backend,yasm-frontend,yasmctl,yasm-proxy-odbc,yasm-integration,yasm-geocoding,yasm-data,yasm-gotenberg
          environments: dev,staging,prod
          environment-links: https://dev-yasm.prodyna.com,https://staging-yasm.prodyna.com,https://yasm.prodyna.com
          verbose: 1
          github-token: ${{ secrets.OVERVIEW_GITHUB_TOKEN }}
          title: "YASM Deployment Overview"
          target-json-File: profile/deployment-overview.json
          target-md-ile: profile/README.md
          template-file: template/default.md

      # Push the generated files
      - name: Commit changes
        run: |
          git config --local user.email "darko@krizic.net"
          git config --local user.name "Deployment Overview"
          git add profile
          git commit -m "Add/update deployment overview"
```
Note that you have to use a GitHub Personal Access Token (PAT) as a secret.

## Required PAT permissions

The PAT should be a Fine Grain PAT that belongs to the target organization and has the following permissions:

| Permission | Access |
| --- | --- |
| Actions | Read |
| Deployments | Read |
| Environments | Read |
| Metadata | Read |
| Pull Requests | Read |
