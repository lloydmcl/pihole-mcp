# Contributing to pihole-mcp

Thanks for your interest in contributing. This document covers how to get started.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for local Pi-hole testing)
- [mise](https://mise.jdx.dev/) (recommended — manages tool versions)
- [just](https://just.systems/) (task runner)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/hexamatic/pihole-mcp.git
cd pihole-mcp

# Install tools and dependencies (one command)
just setup

# Start local Pi-hole for testing
just dev-up

# Run quality checks
just check
```

## Development Workflow

1. **Create a branch** from `main` for your change
2. **Start the dev environment:** `just dev-up` (Pi-hole at http://localhost:8081)
3. **Make your changes** — one logical change per PR
4. **Run quality checks:** `just check` (format, lint, test)
5. **Test against live Pi-hole:** `just integration`
6. **Submit a pull request** with a clear description of the change

## Code Standards

- **Australian English** spelling in all public-facing text (commit messages, PR descriptions, comments, docs). Use "colour", "behaviour", "organisation", "analyse", etc.
- **Go conventions** — follow existing patterns in the codebase. Run `just fmt` before committing.
- **Tool descriptions** should be 15-25 words, front-loaded with purpose.
- **Response formatting** — use `format.` helpers from `internal/format/`. Prefer compact text over Markdown headings.
- **Error handling** — return `mcp.NewToolResultError()` for domain errors, not Go errors.

## Testing

- **Unit tests:** `just test` — uses `httptest.Server` with mocked API responses
- **Integration tests:** `just integration` — requires `just dev-up` running
- **Linting:** `just lint` — uses golangci-lint with strict config (UK spelling enforced)

Every tool handler should have corresponding unit tests in a `_test.go` file.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add pihole_metrics_cache tool
fix: handle empty DNS unit in Docker environments
docs: update README integration guide for Cursor
test: add unit tests for stats handlers
```

## Adding a New Tool

1. Add the Pi-hole API response type to `internal/pihole/types.go`
2. Create the tool handler in the appropriate file under `internal/tools/`
3. Register it in the category's `Register*()` function
4. Add the `Register*()` call to `internal/tools/registry.go` if it's a new category
5. Add unit tests with mocked API responses
6. Run `just check` to verify everything passes
7. Test against live Pi-hole with `just integration`

## Licence

By contributing, you agree that your contributions will be licenced under the MIT Licence.
