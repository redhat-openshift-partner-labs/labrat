# Implementation Plan: `labrat hub managedclusters` Command

## Overview

Implement a new `labrat hub managedclusters` subcommand to list ACM ManagedCluster custom resources from the hub cluster. The command will display cluster status information in table or JSON format with optional filtering.

## Command Specification

**Usage**: `labrat hub managedclusters [flags]`

**Flags**:
- `--output, -o`: Output format (table|json), default: table
- `--status`: Filter by status (Ready|NotReady|Unknown), optional

**Output Fields** (table format):
- NAME: Cluster name
- STATUS: Overall status (Ready/NotReady/Unknown)
- AVAILABLE: ManagedClusterConditionAvailable status (True/False/Unknown)

**Example Output**:
```
NAME                STATUS      AVAILABLE
cluster-east-1      Ready       True
cluster-west-1      NotReady    False
cluster-central     Unknown     Unknown
```

## Architecture

### Package Structure

```
pkg/kube/          - Kubernetes client initialization
pkg/hub/           - ManagedCluster business logic and output formatting
test/helpers/      - Kubernetes test utilities
test/fixtures/     - Sample ManagedCluster YAML files
cmd/labrat/        - Command definition (modify main.go)
```

### Key Components

1. **Kubernetes Client** (`pkg/kube/client.go`)
   - Load kubeconfig from configured path
   - Create dynamic client for CRD access
   - Handle context selection

2. **ManagedCluster Logic** (`pkg/hub/managedclusters.go`)
   - List all ManagedCluster resources
   - Derive status from conditions
   - Apply status-based filtering

3. **Output Formatting** (`pkg/hub/output.go`)
   - Table output with text/tabwriter
   - Pretty-printed JSON output

4. **Command Integration** (`cmd/labrat/main.go`)
   - Add managedclusters subcommand
   - Wire config → client → hub logic → output

## Implementation Steps

### Phase 1: Dependencies (15 min)

Add to `go.mod`:
```go
k8s.io/api v0.31.4
k8s.io/apimachinery v0.31.4
k8s.io/client-go v0.31.4
open-cluster-management.io/api v0.15.0
```

Run: `go get <dependencies> && go mod tidy`

### Phase 2: Kubernetes Client Package (TDD, ~1 hour)

**File**: `pkg/kube/client_suite_test.go`
- Create Ginkgo test suite

**File**: `pkg/kube/client_test.go`
- Test kubeconfig loading (valid/invalid paths)
- Test context selection
- Test client creation

**File**: `pkg/kube/client.go`
```go
type Client struct {
    config  *rest.Config
    dynamic dynamic.Interface
}

func NewClient(kubeconfigPath string, context string) (*Client, error)
func (c *Client) GetDynamicClient() dynamic.Interface
```

**Key Functions**:
- Use `clientcmd.BuildConfigFromFlags()` for kubeconfig loading
- Create dynamic client with `dynamic.NewForConfig()`
- Handle context selection via `clientcmd.NewNonInteractiveDeferredLoadingClientConfig()`

### Phase 3: ManagedCluster Types & Logic (TDD, ~2 hours)

**File**: `pkg/hub/types.go`
```go
type ClusterStatus string
const (
    StatusReady    ClusterStatus = "Ready"
    StatusNotReady ClusterStatus = "NotReady"
    StatusUnknown  ClusterStatus = "Unknown"
)

type ManagedClusterInfo struct {
    Name      string
    Status    ClusterStatus
    Available string
    Message   string
}

type ManagedClusterFilter struct {
    Status ClusterStatus
}
```

**File**: `pkg/hub/managedclusters_suite_test.go`
- Create Ginkgo test suite

**File**: `pkg/hub/managedclusters_test.go`
- Test List() with fake dynamic client
- Test Filter() with sample data
- Test deriveStatus() with table-driven tests covering:
  - Available=True → Ready
  - Available=False → NotReady
  - Available=Unknown → Unknown
  - No conditions → Unknown
  - Unreachable taint present → NotReady

**File**: `pkg/hub/managedclusters.go`
```go
type ManagedClusterClient interface {
    List(ctx context.Context) ([]ManagedClusterInfo, error)
    Filter(clusters []ManagedClusterInfo, filter ManagedClusterFilter) []ManagedClusterInfo
}

func NewManagedClusterClient(dynamicClient dynamic.Interface) ManagedClusterClient
func (m *managedClusterClient) List(ctx context.Context) ([]ManagedClusterInfo, error)
func (m *managedClusterClient) Filter(...) []ManagedClusterInfo
func deriveStatus(cluster *clusterv1.ManagedCluster) ClusterStatus
func getAvailableCondition(cluster *clusterv1.ManagedCluster) (string, string)
```

**List() Implementation**:
1. Define GVR: `cluster.open-cluster-management.io/v1, Resource=managedclusters`
2. Call `dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})`
3. Unmarshal each item to `clusterv1.ManagedCluster`
4. Extract name, derive status, get available condition
5. Return `[]ManagedClusterInfo`

**deriveStatus() Logic**:
1. Check for `cluster.open-cluster-management.io/unreachable` taint → NotReady
2. Find condition with type "ManagedClusterConditionAvailable"
3. If status="True" → Ready
4. If status="False" → NotReady
5. Otherwise → Unknown

### Phase 4: Output Formatting (TDD, ~1 hour)

**File**: `pkg/hub/output_test.go`
- Test table output formatting
- Test JSON output structure
- Test empty cluster list handling

**File**: `pkg/hub/output.go`
```go
type OutputFormat string
const (
    OutputFormatTable OutputFormat = "table"
    OutputFormatJSON  OutputFormat = "json"
)

type OutputWriter struct {
    format OutputFormat
    writer io.Writer
}

func NewOutputWriter(format OutputFormat, writer io.Writer) *OutputWriter
func (o *OutputWriter) Write(clusters []ManagedClusterInfo) error
```

**Table Output**:
- Use `text/tabwriter` for column alignment
- Header: "NAME\tSTATUS\tAVAILABLE\n"
- Rows: "%s\t%s\t%s\n"

**JSON Output**:
- Use `json.MarshalIndent()` for pretty printing
- Indent with 2 spaces

### Phase 5: Command Integration (~45 min)

**File**: `cmd/labrat/main.go`

Add after `hubStatusCmd` (around line 38):
```go
hubManagedClustersCmd := &cobra.Command{
    Use:   "managedclusters",
    Short: "List ACM managed clusters",
    Long:  `List all managed clusters from the ACM hub with status information.`,
    RunE: func(cmd *cobra.Command, _ []string) error {
        // 1. Get flags
        configPath, _ := cmd.Flags().GetString("config")
        outputFormat, _ := cmd.Flags().GetString("output")
        statusFilter, _ := cmd.Flags().GetString("status")

        // 2. Load config
        cfg, err := config.Load(os.ExpandEnv(configPath))
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }

        // 3. Create Kubernetes client
        kubeClient, err := kube.NewClient(cfg.GetHubKubeconfig(), cfg.Hub.Context)
        if err != nil {
            return fmt.Errorf("failed to create kubernetes client: %w", err)
        }

        // 4. Create ManagedCluster client
        mcClient := hub.NewManagedClusterClient(kubeClient.GetDynamicClient())

        // 5. List clusters
        ctx := context.Background()
        clusters, err := mcClient.List(ctx)
        if err != nil {
            return fmt.Errorf("failed to list managed clusters: %w", err)
        }

        // 6. Apply filter if specified
        if statusFilter != "" {
            filter := hub.ManagedClusterFilter{
                Status: hub.ClusterStatus(statusFilter),
            }
            clusters = mcClient.Filter(clusters, filter)
        }

        // 7. Output results
        output := hub.NewOutputWriter(hub.OutputFormat(outputFormat), os.Stdout)
        if err := output.Write(clusters); err != nil {
            return fmt.Errorf("failed to write output: %w", err)
        }

        return nil
    },
}

hubManagedClustersCmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
hubManagedClustersCmd.Flags().String("status", "", "Filter by status (Ready|NotReady|Unknown)")

hubCmd.AddCommand(hubStatusCmd, hubManagedClustersCmd)
```

Add imports:
```go
import (
    "context"
    "os"

    "github.com/redhat-openshift-partner-labs/labrat/internal/config"
    "github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
    "github.com/redhat-openshift-partner-labs/labrat/pkg/kube"
)
```

### Phase 6: Test Fixtures & Helpers (~30 min)

**File**: `test/fixtures/managedcluster_ready.yaml`
- Sample ManagedCluster with Available=True

**File**: `test/fixtures/managedcluster_notready.yaml`
- Sample ManagedCluster with Available=False and unreachable taint

**File**: `test/helpers/kubernetes.go`
```go
func CreateTestManagedCluster(name string, available string) *clusterv1.ManagedCluster
func LoadManagedClusterFromFile(path string) (*clusterv1.ManagedCluster, error)
```

### Phase 7: Testing & Quality (~1 hour)

**Commands**:
```bash
# Run tests with watch mode during development
task test:watch

# Run full test suite
task test:all

# Check coverage (target: 80%+)
task test:coverage

# Run linter
task lint

# Format code
task fmt

# Full quality check
task check
```

**Coverage Targets**:
- pkg/kube: 90%+
- pkg/hub: 90%+
- cmd/labrat: 85%+
- Overall: 80%+

## Critical Files

The following files are essential for this implementation:

1. **`/home/mrhillsman/Development/goland/labrat/pkg/kube/client.go`**
   - Foundation for Kubernetes interactions
   - Creates clients from kubeconfig
   - Must be implemented first

2. **`/home/mrhillsman/Development/goland/labrat/pkg/hub/managedclusters.go`**
   - Core business logic
   - Status derivation algorithm
   - ManagedCluster listing and filtering

3. **`/home/mrhillsman/Development/goland/labrat/pkg/hub/output.go`**
   - User-facing output formatting
   - Table and JSON modes

4. **`/home/mrhillsman/Development/goland/labrat/cmd/labrat/main.go`**
   - Command integration
   - Wires all components together

5. **`/home/mrhillsman/Development/goland/labrat/pkg/hub/managedclusters_test.go`**
   - Comprehensive test coverage
   - Drives TDD development

## Status Derivation Algorithm

Based on the provided ManagedCluster manifest:

1. **Check for unreachable taint**:
   - If `spec.taints[]` contains taint with key `cluster.open-cluster-management.io/unreachable`
   - → Status: NotReady

2. **Check ManagedClusterConditionAvailable**:
   - Find condition in `status.conditions[]` where `type: "ManagedClusterConditionAvailable"`
   - If `status: "True"` → Status: Ready
   - If `status: "False"` → Status: NotReady
   - If `status: "Unknown"` → Status: Unknown

3. **Default**:
   - If no conditions present → Status: Unknown

## Verification Steps

After implementation, verify with the following tests:

### Manual Testing

```bash
# Build
task build

# Test basic listing
./bin/labrat hub managedclusters

# Test JSON output
./bin/labrat hub managedclusters --output json

# Test filtering
./bin/labrat hub managedclusters --status Ready
./bin/labrat hub managedclusters --status NotReady
./bin/labrat hub managedclusters --status Unknown

# Test with verbose
./bin/labrat hub managedclusters --verbose

# Test error handling (invalid config)
./bin/labrat hub managedclusters --config /nonexistent/path
```

### Expected Behavior

1. **No clusters**: Should display empty table or `[]` for JSON
2. **Multiple clusters**: Should display sorted by name
3. **Invalid kubeconfig**: Should show helpful error message
4. **Invalid status filter**: Should show validation error
5. **Connection errors**: Should propagate Kubernetes API errors clearly

### Success Criteria

- ✅ All unit tests pass (`task test`)
- ✅ Coverage ≥ 80% (`task test:coverage`)
- ✅ Linter passes (`task lint`)
- ✅ Command works against real ACM hub cluster
- ✅ Table output is properly aligned
- ✅ JSON output is valid and pretty-printed
- ✅ Filtering works correctly
- ✅ Error messages are clear and actionable
- ✅ Help text is accurate (`labrat hub managedclusters --help`)

## Dependencies Summary

Add to `go.mod`:
```
k8s.io/api v0.31.4
k8s.io/apimachinery v0.31.4
k8s.io/client-go v0.31.4
open-cluster-management.io/api v0.15.0
```

Key import paths:
```go
"k8s.io/client-go/tools/clientcmd"
"k8s.io/client-go/rest"
"k8s.io/client-go/dynamic"
"k8s.io/apimachinery/pkg/runtime/schema"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
clusterv1 "open-cluster-management.io/api/cluster/v1"
```

## Estimated Timeline

- Phase 1 (Dependencies): 15 minutes
- Phase 2 (Kube Client): 1 hour
- Phase 3 (ManagedCluster Logic): 2 hours
- Phase 4 (Output Formatting): 1 hour
- Phase 5 (Command Integration): 45 minutes
- Phase 6 (Fixtures & Helpers): 30 minutes
- Phase 7 (Testing & Quality): 1 hour

**Total**: ~6.5 hours (following TDD approach with comprehensive testing)
