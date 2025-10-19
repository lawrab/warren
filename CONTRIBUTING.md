# Contributing to Warren

Thank you for your interest in contributing to Warren! This document explains our development process and how to contribute.

## Table of Contents
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Git Hooks](#git-hooks)

## Development Setup

### Prerequisites
- Go 1.23 or later
- GTK4 development libraries
- golangci-lint
- Git

### Using Nix (Recommended)
```bash
# Enter development environment with all dependencies
nix develop

# Or use direnv
echo "use flake" > .envrc && direnv allow
```

### Manual Setup
```bash
# Install GTK4 (Ubuntu/Debian)
sudo apt-get install libgtk-4-dev pkg-config

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Clone and setup
git clone https://github.com/lawrab/warren.git
cd warren
go mod download

# Setup git hooks (recommended)
bash .githooks/setup.sh
```

## Development Workflow

### Branch Strategy
Warren uses a feature branch workflow with protected main branch:

```
main (protected)
  ↓ create branch
feature/your-feature
  ↓ commits
  ↓ push
Pull Request
  ↓ CI passes + review
  ↓ merge
main (updated)
```

### Branch Naming
- `feature/description` - New features
- `bugfix/description` - Bug fixes
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Test additions/changes

### Creating a Feature Branch
```bash
# Start from up-to-date main
git checkout main
git pull

# Create your feature branch
git checkout -b feature/my-feature

# Make changes, commit often
git add .
git commit -m "Add feature: description"

# Push when ready
git push -u origin feature/my-feature
```

## Code Standards

### Format and Style
```bash
# Format code (required before commit)
go fmt ./...

# Run linter
golangci-lint run

# Check for issues
go vet ./...
```

### Code Quality Rules
- **Always handle errors** - Never ignore error returns
- **Write tests** - Add tests for new features
- **Keep functions small** - Max ~50 lines when possible
- **Document exported functions** - Use godoc format
- **No panic() for expected errors** - Return errors instead

### Example Code
```go
// GoodExample demonstrates proper error handling and documentation.
// It returns the processed result or an error if processing fails.
func GoodExample(input string) (string, error) {
    if input == "" {
        return "", fmt.Errorf("input cannot be empty")
    }

    result, err := process(input)
    if err != nil {
        return "", fmt.Errorf("failed to process: %w", err)
    }

    return result, nil
}
```

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Run specific package
go test ./internal/fileops/ -v
```

### Writing Tests
- Use table-driven tests
- Test both success and error cases
- Use `t.TempDir()` for filesystem tests
- Keep tests focused and readable

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Example(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wanted error: %v, got: %v", tt.wantErr, err)
            }
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

## Pull Request Process

### Before Creating a PR
1. ✅ All tests pass locally
2. ✅ Code is formatted (`go fmt ./...`)
3. ✅ Linter passes (`golangci-lint run`)
4. ✅ Commits are clean and descriptive
5. ✅ Branch is up-to-date with main

### Creating a PR
```bash
# Push your branch
git push -u origin feature/my-feature

# Create PR (requires gh CLI)
gh pr create --title "Add feature: description" --body "..."

# Or use GitHub web interface
```

### PR Requirements
All PRs must:
- ✅ Pass all CI checks (test, lint, format, build)
- ✅ Have clear description of changes
- ✅ Include tests for new functionality
- ✅ Update documentation if needed
- ✅ Get approval from maintainer

### CI Checks
The following automated checks run on every PR:
- **Test**: Go 1.23 and 1.25, with race detector
- **Lint**: golangci-lint with project configuration
- **Format**: gofmt formatting check
- **Build**: Binary builds and executes

## Git Hooks

### Pre-commit Hook
Warren includes a pre-commit hook that runs:
1. Format check (`gofmt`)
2. Linting (`golangci-lint`)
3. Tests (`go test`)

### Setup
```bash
# One-time setup
bash .githooks/setup.sh
```

### Skipping Hooks
If needed (not recommended):
```bash
git commit --no-verify
```

## Commit Messages

### Format
```
<type>: <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Build/tooling changes

### Examples
```
feat: Add sorting by file size

Implements ascending and descending sort by file size.
Users can toggle with 's' key.

Closes #123

---

fix: Handle permission errors gracefully

Show user-friendly error dialog instead of crashing when
encountering permission denied errors.

---

docs: Update CONTRIBUTING.md with test guidelines

Add examples of table-driven tests and coverage requirements.
```

## Questions?

- Check existing issues: https://github.com/lawrab/warren/issues
- Read documentation in `docs/`
- Review `CLAUDE.md` for project context

## Code of Conduct

Be respectful, constructive, and collaborative. Warren is a learning project - we're all here to learn and improve together.
