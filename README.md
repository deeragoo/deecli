# deecli

[![Build Status](https://github.com/deeragoo/deecli/actions/workflows/ci.yml/badge.svg)](https://github.com/deeragoo/deecli/actions)
[![Latest Release](https://img.shields.io/github/v/release/deeragoo/deecli)](https://github.com/deeragoo/deecli/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org/dl/)
[![GitHub stars](https://img.shields.io/github/stars/deeragoo/deecli?style=social)](https://github.com/deeragoo/deecli/stargazers)


`deecli` is an all-in-one developer shortcut CLI tool designed to simplify common developer tasks with handy shortcuts and integrations.


```bash
# Example: List running Docker containers
deecli docker-ps
```
---

## Features

- **AWS S3 List**: Quickly list your AWS S3 buckets.
- **Docker PS**: List running Docker containers.
- **Git Automation**:
  - Initialize a git repository, commit all files, set remote origin, and push.
  - Create and push git tags.
- **GitHub Integration**:
  - Create GitHub repositories via the GitHub API.
  - Encrypt and securely store GitHub tokens for reuse.
- **CLI Update**: Easily update the `deecli` to the latest version easily.

---

## Installation

### Easy Installation Scripts

To simplify installation, you can use the provided install scripts:

- For Linux/macOS:

  ```bash
  chmod +x install.sh
  ./install.sh
  ```

- For Windows PowerShell:
  ```
  .\install.ps1
  ```

These scripts will copy the deecli binary to ~/.deecli/bin (or %USERPROFILE%\.deecli\bin on Windows) and add that directory to your PATH automatically.

Make sure you download the appropriate binary for your platform and place it in the same directory as the install script before running it.


You can download pre-built binaries for your platform from the [GitHub Releases](https://github.com/deeragoo/deecli/releases/tag/v1.0.10).

Alternatively, you can build from source:

```bash
git clone https://github.com/deeragoo/deecli.git
cd deecli/cmd/deecli
go build -o deecli main.go
```
---
# Usage

```
deecli [command] [args]
```

## Available Commands
| Command            | Description                                              |
|--------------------|----------------------------------------------------------|
| aws-s3-ls          | List AWS S3 buckets (requires AWS CLI configured)        |
| docker-ps          | List running Docker containers                           |
| git-init-push      | Initialize git repo, commit all, set remote, and push    |
| git-tag-push       | Create a git tag and push to origin                      |
| github-create-repo | Create a GitHub repository via API                       |
| encrypt-token      | Encrypt a GitHub token interactively and save securely   |
| decrypt-token      | Decrypt and display the GitHub token                     |
| update             | Update deecli to the latest version                      |


# Examples

## List AWS S3 Buckets
```
deecli aws-s3-ls
```

## List Running Docker Containers
```
deecli docker-ps
```

## Initialize Git Repo and Push
```
deecli git-init-push https://github.com/username/repo.git
```

## Create and Push Git Tag
```
deecli git-tag-push v1.0.0
```

## Create GitHub Repository
Make sure you have a GitHub token set in GH_TOKEN environment variable or encrypted in ~/.secrets.json.

```
deecli github-create-repo my-new-repo --private
```

## Encrypt GitHub Token
```
deecli encrypt-token
```

## Decrypt Github Token
```
deecli decrypt-token
```

## Update deecli
```
deecli update
```

## Github Token Configuration for Github Commands

- GitHub Token: The CLI uses your GitHub token for API operations. You can provide it via the GH_TOKEN environment variable or securely encrypt it using the encrypt-token command. The encrypted token is stored in ~/.secrets.json.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT License

## Author

Deeragoo