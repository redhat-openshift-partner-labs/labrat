# Linting Configuration

This document describes the linting setup for LABRAT.

## Overview

LABRAT uses **golangci-lint**, a comprehensive linting aggregator that runs multiple linters in parallel for fast, thorough code quality checks.

## Quick Start

```bash
# Run linter
task lint

# Auto-fix issues where possible
task lint:fix

# Run with verbose output
task lint:verbose

# Run all quality checks (format, vet, lint, test)
task check
```

## Enabled Linters

### Core Linters (Default)
- **errcheck** - Checks for unchecked errors
- **gosimple** - Suggests code simplifications
- **govet** - Go's official vet tool
- **ineffassign** - Detects ineffectual assignments
- **staticcheck** - Advanced static analysis
- **unused** - Finds unused constants, variables, functions

### Additional Linters
- **gofmt** - Code formatting checks
- **goimports** - Import ordering and formatting
- **misspell** - Finds commonly misspelled words
- **revive** - Fast, configurable linter (golint replacement)
- **gosec** - Security-focused checks
- **unconvert** - Removes unnecessary type conversions
- **unparam** - Finds unused function parameters
- **gocyclo** - Cyclomatic complexity (threshold: 15)
- **dupl** - Code clone detection (threshold: 100 tokens)
- **goconst** - Finds repeated strings for constants
- **gocognit** - Cognitive complexity (threshold: 20)
- **bodyclose** - HTTP response body closure checks
- **nilerr** - Finds incorrect nil error returns
- **prealloc** - Slice preallocation suggestions
- **stylecheck** - Style recommendations
- **whitespace** - Leading/trailing whitespace detection

## Configuration

Configuration is in `.golangci.yml` at the project root.

### Key Settings

**Complexity Thresholds:**
- Cyclomatic complexity: 15
- Cognitive complexity: 20
- Duplicate threshold: 100 tokens

**Constant Detection:**
- Min string length: 3 characters
- Min occurrences: 3

**Test File Exclusions:**
- Test files (`_test.go`) have relaxed rules for:
  - Cyclomatic complexity
  - Unchecked errors
  - Code duplication
  - Security checks
  - Constants

**Dot Imports:**
- Allowed in test files (idiomatic for Ginkgo/Gomega)
- Allowed in test helpers

**Generated Code:**
- Excluded: `mocks/`, `*.pb.go`, `*.gen.go`
- Excluded directories: `bin/`, `vendor/`, `.serena/`, `.idea/`

## Common Issues and Fixes

### Unused Parameters

**Issue:**
```go
func handler(cmd *cobra.Command, args []string) {
    fmt.Println("Hello")
}
```

**Fix:**
```go
func handler(_ *cobra.Command, _ []string) {
    fmt.Println("Hello")
}
```

### Unchecked Errors

**Issue:**
```go
value, _ := cmd.Flags().GetString("flag")
```

**Fix:**
```go
value, err := cmd.Flags().GetString("flag")
if err != nil {
    return err
}
```

### Security Issues (gosec)

**Issue:**
```go
cmd := exec.Command("sh", "-c", userInput)  // G204: command injection
```

**Fix:**
```go
// Use explicit arguments instead of shell interpolation
cmd := exec.Command("program", arg1, arg2)
```

## Auto-Fix

Some issues can be auto-fixed:

```bash
task lint:fix
```

Auto-fixable issues include:
- Code formatting (gofmt)
- Import ordering (goimports)
- Unnecessary conversions (unconvert)
- Whitespace issues

## Ignoring Specific Issues

### Inline Comments

```go
//nolint:unused // will be used in future version command
var version = "0.1.0"

//nolint:gosec // G204 is acceptable here, input is validated
cmd := exec.Command("sh", "-c", validatedInput)
```

### File-Level

```go
//nolint:gocyclo // complex business logic
func complexFunction() {
    // ...
}
```

### Configuration

Add exclusions to `.golangci.yml`:

```yaml
issues:
  exclude-rules:
    - path: specific/file.go
      linters:
        - errcheck
```

## CI Integration

Use `task ci` for CI/CD pipelines:

```bash
task ci
```

This runs:
1. Format check (fails if code is not formatted)
2. `go vet`
3. `golangci-lint`
4. All tests (unit + integration + e2e)

### GitHub Actions Example

```yaml
- name: Run quality checks
  run: task ci
```

## Performance

golangci-lint runs linters in parallel and caches results, making it very fast:

- Initial run: ~2-5 seconds (small codebase)
- Cached runs: ~1-2 seconds
- Only changed files: <1 second

## Updating golangci-lint

```bash
# Update to latest version
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or use the task
task lint:deps
```

## Documentation

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Configuration Reference](https://golangci-lint.run/usage/configuration/)
- [Linters List](https://golangci-lint.run/usage/linters/)

## Status

✅ golangci-lint installed and configured
✅ 24 linters enabled
✅ Test-specific exclusions configured
✅ Auto-fix capability enabled
✅ All code passes linting
✅ Integrated with Taskfile
✅ Ready for CI/CD integration
