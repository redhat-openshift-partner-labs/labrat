# Testing Infrastructure Setup

This document describes the testing infrastructure that has been set up for LABRAT.

## What's Been Configured

### 1. Testing Frameworks Installed
- ✅ **Ginkgo v2** - BDD testing framework
- ✅ **Gomega** - Matcher/assertion library
- ✅ **counterfeiter** - Mock generation tool
- ✅ **gopkg.in/yaml.v3** - YAML parsing (for config)

### 2. Taskfile Commands Added

```bash
# Install testing tools
task test:deps

# Run unit tests (fast, default)
task test
task test:unit

# Run integration tests
task test:integration

# Run E2E tests
task test:e2e

# Run all test types
task test:all

# Generate coverage report
task test:coverage

# Watch mode for TDD
task test:watch

# Run with race detector
task test:race

# Generate mocks
task mocks:generate
```

### 3. Directory Structure Created

```
labrat/
├── internal/config/
│   ├── config.go                    # ✅ Implementation
│   ├── config_test.go               # ✅ BDD tests (12 passing specs)
│   └── config_suite_test.go         # ✅ Ginkgo suite setup
├── test/
│   ├── e2e/
│   │   ├── e2e_suite_test.go        # ✅ E2E suite setup
│   │   └── hub_status_test.go       # ✅ Example E2E test (stubbed)
│   ├── fixtures/
│   │   └── valid_config.yaml        # ✅ Test fixture example
│   ├── helpers/
│   │   └── test_helpers.go          # ✅ Shared test utilities
│   └── README.md                    # ✅ Testing guide
├── coverage/                        # ✅ Coverage reports (gitignored)
├── .gitignore                       # ✅ Updated
└── TESTING.md                       # ✅ This file
```

### 4. Example Test Suite

The `internal/config` package demonstrates:
- ✅ Ginkgo BDD test structure
- ✅ Describe/Context/It blocks
- ✅ BeforeEach/AfterEach setup
- ✅ Table-driven tests with DescribeTable
- ✅ Gomega matchers and assertions
- ✅ TDD approach (tests written first)

**All 12 tests pass!** ✅

## Quick Start: Running Tests

### Run the example tests
```bash
# Run config package tests
ginkgo internal/config/

# Or use task
task test
```

### Start TDD development
```bash
# Watch mode - auto-runs tests on file changes
task test:watch
```

### Generate coverage report
```bash
task test:coverage
# Opens coverage/coverage.html
```

## TDD/BDD Workflow

### 1. Write Test First (RED)
```go
// internal/hub/hub_test.go
Describe("Hub Status", func() {
    It("should return healthy status", func() {
        status, err := hubClient.GetStatus()
        Expect(err).NotTo(HaveOccurred())
        Expect(status).To(Equal(Healthy))
    })
})
```

### 2. Run Test (Should Fail)
```bash
task test:watch  # Auto-runs on changes
# OR
ginkgo internal/hub/
```

### 3. Implement Code (GREEN)
```go
// internal/hub/hub.go
func (h *HubClient) GetStatus() (Status, error) {
    return Healthy, nil
}
```

### 4. Test Passes ✅

### 5. Refactor (While Tests Stay Green)
Improve code quality, extract functions, optimize, etc.

## Coverage Targets

| Component | Target | Status |
|-----------|--------|--------|
| Overall | 80%+ | 🎯 Set as goal |
| internal/config | 80%+ | ✅ 100% (example) |
| pkg/* | 90%+ | 🔜 To be implemented |
| cmd/labrat | 85%+ | 🔜 To be implemented |

## Next Steps for Development

### Implementing New Features with TDD

1. **Create test suite** for the package
   ```bash
   ginkgo bootstrap pkg/hub
   ginkgo generate pkg/hub/status
   ```

2. **Write failing tests** describing desired behavior

3. **Implement minimum code** to pass tests

4. **Refactor** while keeping tests green

5. **Add integration tests** if needed

### Example: Adding Hub Package

```bash
# 1. Create test suite
cd pkg/hub
ginkgo bootstrap

# 2. Write test file
# pkg/hub/hub_test.go

# 3. Start watch mode
task test:watch

# 4. Implement pkg/hub/hub.go

# 5. Tests pass!
```

## Best Practices Enforced

- ✅ Tests run before implementation (TDD)
- ✅ BDD-style readable test descriptions
- ✅ Arrange-Act-Assert pattern
- ✅ Interface-based mocking
- ✅ Separation of test types (unit/integration/e2e)
- ✅ Coverage reporting
- ✅ Race detection capability
- ✅ Comprehensive linting with golangci-lint
- ✅ Code formatting with gofmt
- ✅ Static analysis with go vet

## Code Quality

### Linting with golangci-lint

```bash
# Run linter
task lint

# Auto-fix issues where possible
task lint:fix

# Verbose output
task lint:verbose
```

### Quality Checks

```bash
# Format code
task fmt

# Run vet
task vet

# Run all quality checks (fmt, vet, lint, test)
task check

# CI pipeline checks
task ci
```

The project uses golangci-lint with a comprehensive set of linters including:
- errcheck, gosimple, govet, ineffassign, staticcheck, unused
- gofmt, goimports, misspell, revive, gosec
- gocyclo, dupl, goconst, and more

Configuration is in `.golangci.yml` with sensible defaults and test-specific exclusions.

## Example Test Output

```
Running Suite: Config Suite
========================
Random Seed: 1768517938

Will run 12 of 12 specs
••••••••••••

Ran 12 of 12 Specs in 0.002 seconds
SUCCESS! -- 12 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
```

## Resources

- See `test/README.md` for detailed testing guide
- See comprehensive strategy in project memory: `testing_strategy.md`
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)

---

**Status**: ✅ Testing infrastructure fully configured and operational
**Example Tests**: ✅ 12 passing tests in internal/config
**Ready for**: TDD/BDD development of all packages
