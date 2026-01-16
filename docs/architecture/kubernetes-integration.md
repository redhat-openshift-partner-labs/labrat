# Kubernetes Integration Architecture

## Overview

LABRAT integrates with Kubernetes clusters to manage ACM (Advanced Cluster Management) resources and spoke clusters. This document describes the architecture and design decisions for Kubernetes integration.

## Design Principles

1. **Client Reusability**: Create Kubernetes clients once and reuse across operations
2. **Type Safety**: Use typed clients from official Kubernetes libraries when possible
3. **Error Handling**: Provide clear, actionable error messages for connection and API failures
4. **Configuration**: Support standard kubeconfig files and context selection
5. **Testing**: Enable comprehensive testing with fake clients

## Package Structure

### `pkg/kube/`

The `kube` package provides low-level Kubernetes client initialization and management.

**Responsibilities**:
- Load kubeconfig from file paths
- Handle Kubernetes context selection
- Create REST configurations
- Provide dynamic clients for CRD access
- Abstract client-go complexity

**Key Types**:
```go
type Client struct {
    config  *rest.Config
    dynamic dynamic.Interface
}
```

**Key Functions**:
```go
// NewClient creates a Kubernetes client from kubeconfig
func NewClient(kubeconfigPath string, context string) (*Client, error)

// GetDynamicClient returns the dynamic client for CRD operations
func (c *Client) GetDynamicClient() dynamic.Interface

// GetConfig returns the REST config for creating additional clients
func (c *Client) GetConfig() *rest.Config
```

### `pkg/hub/`

The `hub` package provides high-level ACM hub cluster operations.

**Responsibilities**:
- List and filter ManagedCluster resources
- List and correlate ClusterDeployment (Hive) resources
- Derive cluster status from conditions
- Format output for CLI display
- Abstract ACM and Hive-specific logic

**Key Types**:
```go
type ManagedClusterClient interface {
    List(ctx context.Context) ([]ManagedClusterInfo, error)
    Filter(clusters []ManagedClusterInfo, filter ManagedClusterFilter) []ManagedClusterInfo
}

type ClusterDeploymentClient interface {
    List(ctx context.Context) ([]ClusterDeploymentInfo, error)
    Get(ctx context.Context, name, namespace string) (*ClusterDeploymentInfo, error)
}

type CombinedClusterClient interface {
    List(ctx context.Context) ([]CombinedClusterInfo, error)
}
```

## Client Architecture

### Client Creation Flow

```
User Command
    ↓
Load Config (internal/config)
    ↓
Create Kube Client (pkg/kube)
    ├── Load kubeconfig file
    ├── Select context
    ├── Build REST config
    └── Create dynamic client
    ↓
Create Hub Clients (pkg/hub)
    ├── Receive dynamic client
    ├── Create ManagedCluster client
    ├── Create ClusterDeployment client
    ├── Create Combined client (correlates both)
    └── Prepare for CRD operations
    ↓
Execute Operations
    ├── List ManagedClusters (basic info)
    ├── List ClusterDeployments (platform/power state)
    ├── Correlate resources (combined view)
    ├── Filter results
    └── Format output
```

### Kubeconfig Loading

LABRAT uses the standard Kubernetes `client-go` library for kubeconfig loading:

1. **Path Resolution**:
   - Use path from `hub.kubeconfig` in config.yaml
   - Support environment variable expansion (e.g., `$HOME`)
   - Validate file existence before loading

2. **Context Selection**:
   - Use `hub.context` from config if specified
   - Fall back to current context if not specified
   - Validate context exists in kubeconfig

3. **Configuration Building**:
   ```go
   clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
       &clientcmd.ClientConfigLoadingRules{
           ExplicitPath: kubeconfigPath,
       },
       &clientcmd.ConfigOverrides{
           CurrentContext: context,
       },
   )
   ```

## Dynamic Client vs Typed Client

### Dynamic Client (Current Implementation)

**Pros**:
- Works with any CRD without generated code
- Flexible for multiple resource types
- No dependency on ACM API versions

**Cons**:
- Requires manual unmarshaling
- Less type safety
- More verbose code

**Usage**:
```go
gvr := schema.GroupVersionResource{
    Group:    "cluster.open-cluster-management.io",
    Version:  "v1",
    Resource: "managedclusters",
}

unstructured, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})

// Manual unmarshaling required
var cluster clusterv1.ManagedCluster
err = runtime.DefaultUnstructuredConverter.FromUnstructured(
    item.UnstructuredContent(),
    &cluster,
)
```

### Typed Client (Alternative)

**Pros**:
- Type-safe operations
- Generated from OpenAPI specs
- Cleaner code

**Cons**:
- Requires ACM client-go dependency
- Version coupling with ACM
- Additional dependencies

**Example** (if we migrate):
```go
import clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"

clusterClient := clusterclient.NewForConfigOrDie(kubeConfig)
clusters, err := clusterClient.ClusterV1().
    ManagedClusters().
    List(ctx, metav1.ListOptions{})
```

**Decision**: Start with dynamic client for flexibility; consider typed client if complexity grows.

## ManagedCluster Resource Handling

### Resource Schema

ManagedCluster is a custom resource defined by ACM:

```yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
```

### Key Fields

**Metadata**:
- `metadata.name`: Cluster identifier
- `metadata.labels`: Cloud provider, region, OpenShift version
- `metadata.annotations`: Creation source, additional metadata

**Spec**:
- `spec.hubAcceptsClient`: Whether hub accepts this cluster
- `spec.leaseDurationSeconds`: Heartbeat interval
- `spec.taints`: Conditions affecting scheduling (e.g., unreachable)

**Status**:
- `status.conditions[]`: Array of cluster conditions
- `status.version.kubernetes`: Kubernetes version
- `status.capacity`: Resource capacity information
- `status.allocatable`: Allocatable resources

### Condition Types

| Type | Meaning |
|------|---------|
| `ManagedClusterConditionAvailable` | Cluster is reachable and healthy |
| `ManagedClusterImportSucceeded` | Import process completed successfully |
| `HubAcceptedManagedCluster` | Hub has accepted this cluster |
| `ManagedClusterJoined` | Cluster has joined the hub |

### Status Derivation Algorithm

```
1. Check spec.taints[] for unreachable taint
   ├── Key: "cluster.open-cluster-management.io/unreachable"
   └── If present → NotReady

2. Find ManagedClusterConditionAvailable in status.conditions[]
   ├── status: "True" → Ready
   ├── status: "False" → NotReady
   └── status: "Unknown" → Unknown

3. Default (no conditions) → Unknown
```

**Implementation**:
```go
func deriveStatus(cluster *clusterv1.ManagedCluster) ClusterStatus {
    // Check for unreachable taint
    for _, taint := range cluster.Spec.Taints {
        if taint.Key == "cluster.open-cluster-management.io/unreachable" {
            return StatusNotReady
        }
    }

    // Check ManagedClusterConditionAvailable
    for _, condition := range cluster.Status.Conditions {
        if condition.Type == "ManagedClusterConditionAvailable" {
            switch condition.Status {
            case metav1.ConditionTrue:
                return StatusReady
            case metav1.ConditionFalse:
                return StatusNotReady
            case metav1.ConditionUnknown:
                return StatusUnknown
            }
        }
    }

    return StatusUnknown
}
```

## Dual Resource Model: ManagedCluster + ClusterDeployment

LABRAT implements a dual resource model to provide complete cluster management capabilities. This model correlates two complementary Kubernetes custom resources on the hub cluster.

### Why Two Resources?

**ManagedCluster (ACM)** and **ClusterDeployment (Hive)** serve different purposes:

| Aspect | ManagedCluster (ACM) | ClusterDeployment (Hive) |
|--------|----------------------|--------------------------|
| **Purpose** | Cluster registration and health monitoring | Cluster lifecycle and provisioning |
| **Scope** | Cluster-scoped | Namespaced (namespace = cluster name) |
| **Provides** | Health status, availability, addons | Power state, platform, region, version, credentials |
| **Created When** | Cluster is imported or created | Cluster is provisioned (not for imported clusters) |
| **API Group** | cluster.open-cluster-management.io | hive.openshift.io |

### Resource Correlation

Both resources share the same cluster name, enabling correlation:

```go
// ManagedCluster
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  name: my-cluster

// ClusterDeployment (in namespace with same name)
apiVersion: hive.openshift.io/v1
kind: ClusterDeployment
metadata:
  name: my-cluster
  namespace: my-cluster
```

### ClusterDeployment Resource Handling

#### Resource Schema

```yaml
apiVersion: hive.openshift.io/v1
kind: ClusterDeployment
metadata:
  name: cluster-name
  namespace: cluster-name
  annotations:
    hive.openshift.io/cluster-platform: AWS
spec:
  powerState: Running  # Running, Hibernating, Stopped
  platform:
    aws:
      region: us-east-1
  provisioning:
    imageSetRef:
      name: openshift-v4.14.8
status:
  adminKubeconfigSecretRef:
    name: cluster-name-admin-kubeconfig
  installedTimestamp: "2024-01-15T10:30:00Z"
  powerState: Running
```

#### Key Fields

**Spec**:
- `spec.powerState`: Desired power state (Running, Hibernating, Stopped)
- `spec.platform`: Cloud platform configuration (AWS, Azure, GCP, etc.)
- `spec.platform.<provider>.region`: Cloud region
- `spec.provisioning.imageSetRef.name`: OpenShift version reference

**Status**:
- `status.powerState`: Current power state
- `status.adminKubeconfigSecretRef`: Reference to admin kubeconfig secret
- `status.installedTimestamp`: When cluster was provisioned

**Metadata**:
- `metadata.annotations["hive.openshift.io/cluster-platform"]`: Platform type

### Combined Client Architecture

The `CombinedClusterClient` merges data from both resources:

```go
type CombinedClusterInfo struct {
    // From ManagedCluster
    Name      string
    Status    ClusterStatus
    Available string
    Message   string

    // From ClusterDeployment
    PowerState string
    Platform   string
    Region     string
    Version    string
}

func (c *CombinedClusterClient) List(ctx context.Context) ([]CombinedClusterInfo, error) {
    // 1. List all ManagedClusters
    managedClusters := managedClusterClient.List(ctx)

    // 2. List all ClusterDeployments
    clusterDeployments := clusterDeploymentClient.List(ctx)

    // 3. Correlate by name
    for each managedCluster {
        info := CombinedClusterInfo{Name: managedCluster.Name, ...}

        if deployment := findDeployment(managedCluster.Name) {
            info.PowerState = deployment.PowerState
            info.Platform = deployment.Platform
            info.Region = deployment.Region
            info.Version = deployment.Version
        } else {
            // No ClusterDeployment (imported cluster)
            info.PowerState = "N/A"
            info.Platform = "N/A"
            info.Region = "N/A"
            info.Version = "N/A"
        }
    }
}
```

### Handling Missing ClusterDeployments

Not all ManagedClusters have a corresponding ClusterDeployment:

**Scenarios**:
- **Imported clusters**: Manually imported clusters lack Hive resources
- **Non-cloud platforms**: Bare metal or vSphere clusters may not use Hive
- **Legacy clusters**: Older clusters created before Hive integration

**Graceful Degradation**:
```go
// --wide flag shows "N/A" for missing ClusterDeployment data
NAME                STATUS      AVAILABLE   POWER STATE   PLATFORM   REGION   VERSION
imported-cluster    Ready       True        N/A           N/A        N/A      N/A
hive-cluster        Ready       True        Running       AWS        us-e-1   4.14.8
```

### Spoke Kubeconfig Extraction

The `spoke kubeconfig` command uses ClusterDeployment to access spoke credentials:

```
1. Find ClusterDeployment for cluster
   ↓
2. Extract status.adminKubeconfigSecretRef
   ↓
3. Read secret from same namespace
   ↓
4. Decode kubeconfig data
   ↓
5. Output to file or stdout
```

**Implementation**:
```go
func ExtractKubeconfig(clusterName string) ([]byte, error) {
    // 1. Get ClusterDeployment
    deployment := clusterDeploymentClient.Get(ctx, clusterName, clusterName)

    // 2. Get secret reference
    secretName := deployment.Status.AdminKubeconfigSecretRef.Name

    // 3. Read secret
    secret := secretClient.Get(ctx, secretName, clusterName)

    // 4. Extract kubeconfig data
    return secret.Data["kubeconfig"], nil
}
```

## Testing Strategy

### Unit Testing with Fake Clients

Use `client-go/dynamic/fake` for unit tests:

```go
import (
    "k8s.io/apimachinery/pkg/runtime"
    dynamicfake "k8s.io/client-go/dynamic/fake"
)

scheme := runtime.NewScheme()
clusterv1.AddToScheme(scheme)

fakeDynamic := dynamicfake.NewSimpleDynamicClient(scheme, testClusters...)
client := NewManagedClusterClient(fakeDynamic)

clusters, err := client.List(context.Background())
```

### Integration Testing

For integration tests with real Kubernetes:

1. Use kind or k3s for test clusters
2. Install ACM operator or create ManagedCluster CRDs
3. Create test ManagedCluster resources
4. Run actual client operations

### Test Fixtures

Store sample ManagedCluster YAML in `test/fixtures/`:
- `managedcluster_ready.yaml`
- `managedcluster_notready.yaml`
- `managedcluster_unknown.yaml`
- `managedcluster_unreachable_taint.yaml`

## Error Handling

### Error Categories

1. **Configuration Errors**:
   - Invalid kubeconfig path
   - Missing required fields
   - Malformed YAML

2. **Connection Errors**:
   - Network unreachable
   - DNS resolution failures
   - TLS certificate errors

3. **Authentication Errors**:
   - Invalid credentials
   - Expired tokens
   - Insufficient RBAC permissions

4. **API Errors**:
   - CRD not installed
   - API server errors
   - Resource not found

### Error Propagation

```go
// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to list managed clusters: %w", err)
}

// Provide actionable messages
if os.IsNotExist(err) {
    return nil, fmt.Errorf("kubeconfig not found at %s: ensure the path is correct", kubeconfigPath)
}
```

## Performance Considerations

### Caching

- **Current**: No caching; fetch on every command execution
- **Future**: Consider caching for watch operations or frequent polling

### Pagination

- **Current**: List all clusters in single request
- **Future**: Implement pagination for fleets with 100+ clusters
  ```go
  listOptions := metav1.ListOptions{
      Limit:    100,
      Continue: continueToken,
  }
  ```

### Concurrency

- Single-threaded for simplicity
- Each command execution is independent
- No shared state between commands

## Dependencies

### Core Kubernetes Libraries

```go
k8s.io/client-go v0.31.4    // Kubernetes client library
k8s.io/api v0.31.4          // Kubernetes API types
k8s.io/apimachinery v0.31.4 // API machinery (metav1, schema, etc.)
```

### ACM and Hive API Types

```go
open-cluster-management.io/api v0.15.0  // ManagedCluster CRD types
github.com/openshift/hive/apis v1.2.0   // ClusterDeployment CRD types
```

### Import Paths

```go
import (
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/dynamic"
    "k8s.io/apimachinery/pkg/runtime/schema"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    corev1 "k8s.io/api/core/v1"

    clusterv1 "open-cluster-management.io/api/cluster/v1"
    hivev1 "github.com/openshift/hive/apis/hive/v1"
)
```

## Security Considerations

### Kubeconfig Handling

- Never log kubeconfig contents
- Support restrictive file permissions (0600)
- Don't store credentials in memory longer than necessary

### RBAC Requirements

Users need the following permissions on the hub cluster:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: labrat-hub-reader
rules:
# ManagedCluster access (cluster-scoped)
- apiGroups: ["cluster.open-cluster-management.io"]
  resources: ["managedclusters"]
  verbs: ["get", "list", "watch"]

# ClusterDeployment access (namespaced)
- apiGroups: ["hive.openshift.io"]
  resources: ["clusterdeployments"]
  verbs: ["get", "list"]

# Secret access for kubeconfig extraction (namespaced)
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
  # Note: In practice, limit secret access to specific namespaces
  # via RoleBindings rather than ClusterRoleBinding
```

**Important**: For security, use namespace-specific RoleBindings for ClusterDeployment and secret access rather than granting cluster-wide permissions.

### TLS Verification

- Always verify server certificates by default
- Support `insecure-skip-tls-verify` only via kubeconfig (not in LABRAT config)

## Future Enhancements

### Implemented ✅

1. **ClusterDeployment Integration**: Correlate ManagedCluster with ClusterDeployment for platform/power state info
2. **Wide Output Mode**: Enhanced cluster listing with platform details
3. **Kubeconfig Extraction**: Extract spoke admin kubeconfig from hub secrets
4. **Combined Client**: Unified view of ManagedCluster + ClusterDeployment data

### Planned

1. **Watch Mode**: Real-time cluster status updates
2. **Cluster Details**: Detailed view of single cluster (`spoke get`)
3. **Power State Management**: Hibernate/resume clusters via ClusterDeployment
4. **Status History**: Track status changes over time
5. **Cluster Provisioning**: Create new spoke clusters via Hive

### Under Consideration

1. **Prometheus Integration**: Pull metrics from managed clusters
2. **Event Streaming**: Display cluster events
3. **Batch Operations**: Apply operations to multiple clusters
4. **Custom Conditions**: User-defined status indicators
5. **ClusterClaim Integration**: Support for cluster pool management

## References

- [Kubernetes client-go Documentation](https://github.com/kubernetes/client-go)
- [ACM ManagedCluster API](https://open-cluster-management.io/concepts/managedcluster/)
- [OpenShift Hive Documentation](https://github.com/openshift/hive/blob/master/docs/using-hive.md)
- [Hive ClusterDeployment API](https://github.com/openshift/hive/blob/master/docs/clusterdeployment.md)
- [Dynamic Client Guide](https://ymqytw.github.io/kubernetes/dynamic-client/)
- [ACM Architecture](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/)
