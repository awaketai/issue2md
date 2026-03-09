# issue2md

A CLI tool that converts GitHub Issues, Pull Requests, and Discussions into well-formatted Markdown documents.

## Features

- **Issue / PR / Discussion** — Automatically detects resource type from URL
- **Full content capture** — Title, metadata, body, and all comments (including PR Review Comments)
- **Reactions support** — Optionally include emoji reaction stats
- **Auto-pagination** — Fetches all comments regardless of count
- **File output** — Write to stdout or save to file with auto-generated filenames

## Installation

```bash
# Clone and build
git clone https://github.com/awaketai/issue2md.git
cd issue2md
make build
```

The binary will be placed at `./bin/issue2md`.

### Requirements

- Go >= 1.24

## Usage

```bash
# Output to stdout
issue2md <github-url>

# Save to file
issue2md <github-url> -o output.md

# Save to directory (auto-generates filename)
issue2md <github-url> -o /path/to/dir/

# Include reaction stats
issue2md <github-url> --with-reactions

# Enable debug logging
issue2md <github-url> --verbose

# Combine options
issue2md <github-url> --with-reactions -o out.md --verbose
```

### Supported URL Patterns

| Resource   | URL Pattern                                                  |
|------------|--------------------------------------------------------------|
| Issue      | `https://github.com/{owner}/{repo}/issues/{number}`          |
| PR         | `https://github.com/{owner}/{repo}/pull/{number}`            |
| Discussion | `https://github.com/{owner}/{repo}/discussions/{number}`     |

### CLI Options

| Option              | Description                          | Default |
|---------------------|--------------------------------------|---------|
| `<url>`             | GitHub Issue/PR/Discussion URL (required) | —       |
| `-o, --output`      | Output file path or directory        | stdout  |
| `--with-reactions`  | Show reaction stats on body and comments | false   |
| `--verbose`         | Print debug logs to stderr           | false   |

### Auto-generated Filenames

When `-o` points to a directory, filenames follow this convention:

```
{owner}_{repo}_issue_{number}.md
{owner}_{repo}_pr_{number}.md
{owner}_{repo}_discussion_{number}.md
```

## Authentication

Set the `GITHUB_TOKEN` environment variable with a GitHub Personal Access Token:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
issue2md https://github.com/golang/go/issues/12345
```

- **Without token**: Unauthenticated access (public repos only, 60 requests/hour)
- **With token**: 5,000 requests/hour

## Output Example

```markdown
# [Issue] Fix login timeout (#12345)

| Field | Value |
|-------|-------|
| **State** | Open |
| **Author** | @octocat |
| **Created** | 2024-01-15T10:30:00Z |
| **Labels** | bug, priority/high |
| **Assignees** | @dev1, @dev2 |
| **Milestone** | v2.0 |
| **Linked PRs** | #456, #789 |

---

When attempting to login with SSO credentials, the request times out
after 30 seconds.

> 👍 5 | ❤️ 2

---

## Comments (2)

### @user1 — 2024-01-16T08:00:00Z

I can reproduce this issue.

---

### @user2 — 2024-01-17T14:30:00Z

Fixed in #456.

---
```

PR Review Comments include file path and line number in the header:

```markdown
### @reviewer — 2024-01-18T10:05:00Z `internal/auth/sso.go#L42`

Should we make the max retry count configurable?
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | Input error (invalid URL, unsupported host, unknown resource type) |
| 2    | Runtime error (API error, network error, auth failure) |

## Development

```bash
# Run all tests
make test

# Build CLI binary
make build

# Build web server (future)
make web

# Clean build artifacts
make clean
```

## License

MIT
