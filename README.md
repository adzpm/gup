# glup

Command-line utility for cloning all accessible GitLab repositories.

## Description

`glup` connects to a GitLab instance via API, retrieves all accessible projects, and clones them to a local directory.

The tool preserves the original directory structure from GitLab
(e.g., `backend/tests/auth` is cloned to `target-dir/backend/tests/auth`).

## Installation

### go

```bash
go install github.com/adzpm/glup@latest
```

## Usage

```
glup [global options] clone [options] [directory]
```

### Global Options

- `--gitlab-host <host>` - GitLab host (e.g., `gitlab.com`). Can be set via `GITLAB_HOST` environment variable.
- `--gitlab-user <user>` - GitLab username. Can be set via `GITLAB_USER` environment variable.
- `--gitlab-token <token>` - GitLab access token. Can be set via `GITLAB_TOKEN` environment variable.

### Clone Command Options

- `--group <group>` - Clone repositories only from the specified group. Group path can include subgroups (e.g.,
  `backend/tests/`).

### Arguments

- `[directory]` - Target directory for cloning. If not specified, uses the current working directory.

## Authentication

Authentication is performed in the following order:

1. **Command-line flags** - If `--gitlab-host`, `--gitlab-user`, and `--gitlab-token` are provided, they are used.
2. **Environment variables** - `GITLAB_HOST`, `GITLAB_USER`, `GITLAB_TOKEN` are checked.
3. **`.netrc` file** - If credentials are not provided via flags or environment variables, the tool attempts to load
   them from `~/.netrc`.

### .netrc Format

The tool searches for machine entries containing "gitlab" in the name (case-insensitive). The first matching entry is
used. If multiple entries are found, a warning is displayed and the first one is used.

Example `.netrc` entry:

```
machine gitlab.com
login username
password access_token
```

The `machine` name is used as the GitLab host. If empty, defaults to `gitlab.com`.

## Exit Codes

- `0` - Success
- Non-zero - Error occurred

## Limitations

- Only supports HTTP/HTTPS cloning (not SSH).
- Requires GitLab API access token with appropriate permissions.
- Does not handle repository updates (only clones if directory doesn't exist or is not a git repository).
- No filtering by project visibility, archived status, or other attributes beyond group membership.
