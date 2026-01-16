# LABRAT Testing Guide

This directory contains testing utilities, fixtures, and end-to-end tests for LABRAT.

## Directory Structure

```
test/
├── e2e/              # End-to-end tests (require real cluster)
├── fixtures/         # Test data files (configs, manifests, mock responses)
├── helpers/          # Shared test utilities and helper functions
└── README.md         # This file
```

## Test Organization

### Unit Tests
Located alongside the code they test (e.g., `pkg/hub/hub_test.go`):
- Fast, isolated tests
- Use mocks for external dependencies
- Run by default: `task test:unit` or `go test ./...`

### Integration Tests
Located in package directories with `//go:build integration` tag:
- Test interaction between components
- May use fake Kubernetes clients
- Run with: `task test:integration`

### E2E Tests
Located in `test/e2e/`:
- Test complete user workflows
- Require real or kind/k3s cluster
- Run with: `task test:e2e`

## Running Tests

```bash
# Run unit tests (default, fast)
task test
task test:unit

# Run with watch mode (TDD)
task test:watch

# Run integration tests
task test:integration

# Run E2E tests
task test:e2e

# Run all tests
task test:all

# Generate coverage report
task test:coverage

# Run with race detector
task test:race
```

## Writing Tests

### BDD Style with Ginkgo

```go
var _ = Describe("Feature Name", func() {
    var (
        // declare variables
    )

    BeforeEach(func() {
        // setup before each test
    })

    AfterEach(func() {
        // cleanup after each test
    })

    Context("when condition X", func() {
        It("should do Y", func() {
            // Arrange
            // Act
            // Assert with Gomega matchers
            Expect(result).To(Equal(expected))
        })
    })
})
```

### Table-Driven Tests

```go
DescribeTable("description",
    func(input Type, expected Type) {
        result := FunctionUnderTest(input)
        Expect(result).To(Equal(expected))
    },
    Entry("case 1", input1, expected1),
    Entry("case 2", input2, expected2),
)
```

## Using Fixtures

Load test fixtures from `test/fixtures/`:

```go
import "github.com/redhat-openshift-partner-labs/labrat/test/helpers"

configPath := filepath.Join("test", "fixtures", "valid_config.yaml")
cfg, err := config.Load(configPath)
```

## Using Helpers

Common test utilities are in `test/helpers/`:

```go
import "github.com/redhat-openshift-partner-labs/labrat/test/helpers"

configPath := helpers.CreateTempConfigFile(yamlContent)
defer helpers.CleanupTempDir(configPath)
```

## Mocking

Generate mocks with counterfeiter:

```go
//go:generate counterfeiter -o mocks/hub_client.go . HubClient

type HubClient interface {
    GetStatus() (*Status, error)
}
```

Then run: `task mocks:generate`

## Coverage Goals

- Overall: 80%+ coverage
- Critical paths: 90%+
- CLI handlers: 85%+

Check coverage: `task test:coverage`
