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

| Paramter | Environment | Required | Default | Example                                                                         | Description                                                                   |
| --- | --- |-------|---------|---------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| --github-token | GITHUB_TOKEN | true  | -       | -                                                                               | The GitHub Personal Access Token (PAT)                                        |
| --environments | ENVIRONMENTS | true  | -       | dev,staging,prod                                                                | Environments to query. Comma separated list.                                  |
| --organization | ORGANIZATION | true  | -       | myorga                                                                          | The GitHub Organization to query for repositories.                            |
| --repositories | REPOSITORIES | true  | -       | frontend,backend                                                                | Repositories to query. Comma separated list.                                  |
| --target-repository | TARGET_REPOSITORY | true  | -       | .github                                                                         | The target repository to commit the result to.                                |
| --target-repository-file | TARGET_REPOSITORY_FILE | true  | -       | profile/README.md | The target repository file to commit the result to.                             |
| --verbose | VERBOSE | false  | 1       | 0 | Verbosity level, 0=info, 1=debug. Overrides the environment variable VERBOSE. |

## Use as GitHub Action

TODO

## Required PAT permissions

The PAT should be a Fine Grain PAT that belongs to the target organization and has the following permissions:

| Permission | Access |
| --- | --- |
| Actions | Read |
| Contents | Read & Write |
| Deployments | Read |
| Environments | Read |
| Metadata | Read |
| Pull Requests | Read |
