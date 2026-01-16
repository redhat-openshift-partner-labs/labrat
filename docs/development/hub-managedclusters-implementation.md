# Hub ManagedClusters Implementation Guide

## Overview

This document provides development guidance for implementing the `labrat hub managedclusters` command. For the detailed implementation plan, see [`.claude/plans/hub-managedclusters-implementation.md`](../../.claude/plans/hub-managedclusters-implementation.md).

## Quick Start

### Prerequisites

- Go 1.25+
- Task installed
- Access to an ACM hub cluster with kubeconfig

### Development Workflow

```bash
# 1. Add dependencies
go get k8s.io/client-go@v0.31.4
go get k8s.io/api@v0.31.4
go get k8s.io/apimachinery@v0.31.4
go get open-cluster-management.io/api@v0.15.0
go mod tidy

# 2. Start TDD watch mode
task test:watch

# 3. Implement following the phase order (see plan)
# Phase 2: pkg/kube (tests first, then implementation)
# Phase 3: pkg/hub (tests first, then implementation)
# Phase 4: pkg/hub/output (tests first, then implementation)
# Phase 5: cmd/labrat/main.go (wire everything together)

# 4. Run quality checks
task check

# 5. Manual testing
task build
./bin/labrat hub managedclusters --help
```

## Implementation Phases

Follow the phases defined in the [implementation plan](../../.claude/plans/hub-managedclusters-implementation.md):

1. **Phase 1**: Dependencies (15 min)
2. **Phase 2**: Kubernetes Client Package (1 hour, TDD)
3. **Phase 3**: ManagedCluster Types & Logic (2 hours, TDD)
4. **Phase 4**: Output Formatting (1 hour, TDD)
5. **Phase 5**: Command Integration (45 min)
6. **Phase 6**: Test Fixtures & Helpers (30 min)
7. **Phase 7**: Testing & Quality (1 hour)

Total estimated time: ~6.5 hours

## Test-Driven Development

This implementation follows strict TDD principles:

### Red-Green-Refactor Cycle

For each component:

1. **RED**: Write failing test first
   ```bash
   # Test will fail (function doesn't exist yet)
   task test:watch
   ```

2. **GREEN**: Write minimal code to pass test
   ```bash
   # Test passes with basic implementation
   ```

3. **REFACTOR**: Improve code while tests pass
   ```bash
   # Tests still pass after refactoring
   ```

### Test File Creation Order

Create test files BEFORE implementation files:

```bash
# Example for pkg/kube
1. pkg/kube/client_suite_test.go      # Ginkgo suite setup
2. pkg/kube/client_test.go            # Tests (RED)
3. pkg/kube/client.go                 # Implementation (GREEN)
```

## Package Implementation Order

### 1. pkg/kube (Foundation)

**Must implement first** - all other packages depend on this.

**Files**:
- `client_suite_test.go` - Ginkgo suite
- `client_test.go` - Tests for kubeconfig loading and client creation
- `client.go` - Kubernetes client wrapper

**Key Test Cases**:
- Valid kubeconfig loading
- Invalid kubeconfig path handling
- Context selection
- Dynamic client creation

**Coverage Target**: 90%+

### 2. pkg/hub/types.go (Data Structures)

**No tests needed** - just type definitions.

Define:
- `ClusterStatus` constants
- `ManagedClusterInfo` struct
- `ManagedClusterFilter` struct
- `OutputFormat` constants

### 3. pkg/hub/managedclusters.go (Core Logic)

**Critical component** - contains status derivation algorithm.

**Files**:
- `managedclusters_suite_test.go` - Ginkgo suite
- `managedclusters_test.go` - Comprehensive tests
- `managedclusters.go` - Implementation

**Key Test Cases**:
- List with fake dynamic client
- Filter by status
- Status derivation (table-driven tests):
  - Available=True → Ready
  - Available=False → NotReady
  - Available=Unknown → Unknown
  - Unreachable taint → NotReady
  - No conditions → Unknown

**Coverage Target**: 95%+

### 4. pkg/hub/output.go (Formatting)

**User-facing** - output must be clean and well-formatted.

**Files**:
- `output_test.go` - Table and JSON formatting tests
- `output.go` - OutputWriter implementation

**Key Test Cases**:
- Table formatting with proper alignment
- JSON pretty-printing
- Empty cluster list handling
- Special characters in cluster names

**Coverage Target**: 90%+

### 5. cmd/labrat/main.go (Command Integration)

**Wires everything together**.

Add to existing main.go:
- New `hubManagedClustersCmd` definition
- Flag definitions (--output, --status)
- Command execution flow
- Error handling

**Testing**:
- Integration tests with mocked clients
- Flag parsing validation
- Error path coverage

**Coverage Target**: 85%+

## Testing Helpers

### Creating Test Fixtures

Use `test/fixtures/` for sample ManagedCluster YAML:

```yaml
# test/fixtures/managedcluster_ready.yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  name: test-cluster-ready
status:
  conditions:
  - type: ManagedClusterConditionAvailable
    status: "True"
```

### Using Fake Dynamic Client

```go
import (
    "k8s.io/apimachinery/pkg/runtime"
    dynamicfake "k8s.io/client-go/dynamic/fake"
    clusterv1 "open-cluster-management.io/api/cluster/v1"
)

// In test
scheme := runtime.NewScheme()
_ = clusterv1.AddToScheme(scheme)

testCluster := &clusterv1.ManagedCluster{
    ObjectMeta: metav1.ObjectMeta{Name: "test-cluster"},
    Status: clusterv1.ManagedClusterStatus{
        Conditions: []metav1.Condition{
            {
                Type:   "ManagedClusterConditionAvailable",
                Status: metav1.ConditionTrue,
            },
        },
    },
}

fakeDynamic := dynamicfake.NewSimpleDynamicClient(scheme, testCluster)
```

## Common Patterns

### Error Handling

Always wrap errors with context:

```go
if err != nil {
    return nil, fmt.Errorf("failed to list managed clusters: %w", err)
}
```

### Table-Driven Tests

Use Ginkgo's DescribeTable for status derivation:

```go
DescribeTable("deriveStatus",
    func(conditions []metav1.Condition, taints []clusterv1.Taint, expected ClusterStatus) {
        cluster := &clusterv1.ManagedCluster{
            Spec:   clusterv1.ManagedClusterSpec{Taints: taints},
            Status: clusterv1.ManagedClusterStatus{Conditions: conditions},
        }
        Expect(deriveStatus(cluster)).To(Equal(expected))
    },
    Entry("Available True", availableTrueCondition, noTaints, StatusReady),
    Entry("Available False", availableFalseCondition, noTaints, StatusNotReady),
    Entry("Unreachable taint", anyCondition, unreachableTaint, StatusNotReady),
    // ... more entries
)
```

### Ginkgo Best Practices

```go
Describe("Feature", func() {
    var (
        client    ManagedClusterClient
        fakeDynamic dynamic.Interface
    )

    BeforeEach(func() {
        // Setup before each test
        fakeDynamic = createFakeDynamicClient()
        client = NewManagedClusterClient(fakeDynamic)
    })

    AfterEach(func() {
        // Cleanup after each test
    })

    Context("when clusters exist", func() {
        It("should list all clusters", func() {
            // Arrange
            // Act
            // Assert
        })
    })
})
```

## Quality Gates

Before committing:

```bash
# 1. All tests pass
task test:all

# 2. Coverage meets targets
task test:coverage
# Overall: 80%+
# pkg/kube: 90%+
# pkg/hub: 90%+

# 3. Linter passes
task lint

# 4. Code is formatted
task fmt

# 5. All quality checks
task check
```

## Git Workflow

### Branch Strategy

```bash
# Create feature branch
git checkout -b feature/hub-managedclusters

# Commit frequently with clear messages
git commit -m "feat(kube): add kubernetes client with tests"
git commit -m "feat(hub): implement managedcluster listing"
git commit -m "feat(hub): add output formatting"
git commit -m "feat(cmd): integrate managedclusters command"

# Push and create PR
git push origin feature/hub-managedclusters
```

### Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat(scope): description` - New feature
- `fix(scope): description` - Bug fix
- `test(scope): description` - Test additions
- `docs(scope): description` - Documentation
- `refactor(scope): description` - Code refactoring

**Scopes**:
- `kube` - pkg/kube changes
- `hub` - pkg/hub changes
- `cmd` - command changes
- `test` - test infrastructure
- `docs` - documentation

## Debugging Tips

### Enable Verbose Logging

```bash
labrat hub managedclusters --verbose
```

### Inspect Kubernetes Client

```go
// In development, add debug logging
if verbose {
    log.Printf("Kubeconfig path: %s", kubeconfigPath)
    log.Printf("Context: %s", context)
    log.Printf("Using namespace: %s", namespace)
}
```

### Test Individual Packages

```bash
# Test only pkg/kube
ginkgo pkg/kube/

# Test only pkg/hub
ginkgo pkg/hub/

# Test with verbose output
ginkgo -v pkg/hub/

# Focus on specific test
# Add FIt, FDescribe, or FContext in code, then:
ginkgo pkg/hub/
```

### Coverage Analysis

```bash
# Generate coverage report
task test:coverage

# View in browser (opens coverage/coverage.html)
# Or view in terminal:
go tool cover -func=coverage/coverage.out
```

## Troubleshooting

### "Package not found" errors

```bash
# Ensure dependencies are installed
go mod download
go mod tidy
```

### Test failures with fake client

Make sure to register the scheme:

```go
scheme := runtime.NewScheme()
_ = clusterv1.AddToScheme(scheme)
```

### Kubeconfig loading errors

Check file permissions and expansion:

```go
expandedPath := os.ExpandEnv(kubeconfigPath)
```

## Performance Benchmarking

For performance-critical code:

```go
func BenchmarkDeriveStatus(b *testing.B) {
    cluster := createTestCluster()
    for i := 0; i < b.N; i++ {
        _ = deriveStatus(cluster)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./pkg/hub/
```

## Documentation Checklist

After implementation:

- [ ] Update main README.md with new command
- [ ] Command documentation in docs/commands/
- [ ] Architecture documentation in docs/architecture/
- [ ] Update help text in command definition
- [ ] Add usage examples
- [ ] Document error messages
- [ ] Update CHANGELOG (if exists)

## References

- [Implementation Plan](../../.claude/plans/hub-managedclusters-implementation.md) - Detailed phase-by-phase plan
- [Command Documentation](../commands/hub-managedclusters.md) - User-facing docs
- [Architecture Documentation](../architecture/kubernetes-integration.md) - Technical design
- [Testing Strategy](../../TESTING.md) - Project testing guidelines
- [Code Style](../../LINTING.md) - Linting and formatting rules
