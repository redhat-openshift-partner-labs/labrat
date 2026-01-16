# Command: `labrat hub managedclusters`

## Synopsis

List all ACM (Advanced Cluster Management) ManagedCluster resources from the hub cluster.

```bash
labrat hub managedclusters [flags]
```

## Description

The `managedclusters` subcommand queries the ACM hub cluster to retrieve information about all managed clusters in the fleet. It displays cluster status information derived from ManagedCluster custom resource conditions and taints.

This command is useful for:
- Getting an overview of all clusters in the partner lab environment
- Checking cluster health and availability status
- Filtering clusters by their operational status
- Exporting cluster information for automation or reporting

## Flags

### `--output, -o`
**Type**: String
**Default**: `table`
**Allowed Values**: `table`, `json`

Specifies the output format for cluster information.

- `table`: Human-readable table format with columns
- `json`: JSON format for programmatic consumption

**Example**:
```bash
# Table output (default)
labrat hub managedclusters

# JSON output
labrat hub managedclusters --output json
labrat hub managedclusters -o json
```

### `--status`
**Type**: String
**Optional**: Yes
**Allowed Values**: `Ready`, `NotReady`, `Unknown`

Filter clusters by their operational status.

**Example**:
```bash
# Show only ready clusters
labrat hub managedclusters --status Ready

# Show only unreachable/not ready clusters
labrat hub managedclusters --status NotReady

# Show clusters with unknown status
labrat hub managedclusters --status Unknown
```

### Global Flags

Inherited from parent commands:

- `--config, -c`: Path to labrat config file (default: `$HOME/.labrat/config.yaml`)
- `--verbose, -v`: Enable debug logging

## Output

### Table Format

The default table output displays three columns:

| Column | Description |
|--------|-------------|
| NAME | The name of the managed cluster |
| STATUS | Overall cluster status: `Ready`, `NotReady`, or `Unknown` |
| AVAILABLE | The value of the ManagedClusterConditionAvailable condition |

**Example Output**:
```
NAME                      STATUS      AVAILABLE
cluster-east-1            Ready       True
cluster-west-1            NotReady    False
cluster-central           Unknown     Unknown
9831783a-citrixudn        NotReady    Unknown
```

### JSON Format

JSON output provides the same information in machine-readable format:

```json
[
  {
    "name": "cluster-east-1",
    "status": "Ready",
    "available": "True",
    "message": "Managed cluster is available"
  },
  {
    "name": "cluster-west-1",
    "status": "NotReady",
    "available": "False",
    "message": "Registration agent stopped updating its lease."
  },
  {
    "name": "cluster-central",
    "status": "Unknown",
    "available": "Unknown",
    "message": ""
  }
]
```

## Status Derivation

The `STATUS` field is derived from the ManagedCluster resource using the following logic:

### 1. Check for Unreachable Taint
If the cluster has a taint with key `cluster.open-cluster-management.io/unreachable`:
- **Status**: `NotReady`

### 2. Check ManagedClusterConditionAvailable
Find the condition with `type: "ManagedClusterConditionAvailable"`:
- If `status: "True"` → **Status**: `Ready`
- If `status: "False"` → **Status**: `NotReady`
- If `status: "Unknown"` → **Status**: `Unknown`

### 3. Default
If no conditions are present:
- **Status**: `Unknown`

## Examples

### List All Clusters
```bash
labrat hub managedclusters
```

Output:
```
NAME                STATUS      AVAILABLE
cluster-prod-1      Ready       True
cluster-dev-1       Ready       True
cluster-test-1      NotReady    False
```

### Export to JSON
```bash
labrat hub managedclusters --output json > clusters.json
```

### Show Only Ready Clusters
```bash
labrat hub managedclusters --status Ready
```

Output:
```
NAME                STATUS      AVAILABLE
cluster-prod-1      Ready       True
cluster-dev-1       Ready       True
```

### Count Clusters by Status
```bash
# Count ready clusters
labrat hub managedclusters --status Ready --output json | jq 'length'

# Count not ready clusters
labrat hub managedclusters --status NotReady --output json | jq 'length'
```

### Use with Custom Config
```bash
labrat hub managedclusters --config /path/to/custom/config.yaml
```

### Enable Verbose Logging
```bash
labrat hub managedclusters --verbose
```

## Configuration

This command requires a valid LABRAT configuration file with hub cluster settings.

**Required Configuration** (`~/.labrat/config.yaml`):
```yaml
hub:
  kubeconfig: /path/to/hub/kubeconfig
  context: hub-cluster-context  # Optional, uses current context if not specified
  namespace: open-cluster-management  # Optional, defaults to open-cluster-management
```

## Error Handling

### Configuration Errors
```bash
$ labrat hub managedclusters --config /nonexistent/config.yaml
Error: failed to load config: failed to read config file: open /nonexistent/config.yaml: no such file or directory
```

### Kubeconfig Errors
```bash
$ labrat hub managedclusters
Error: failed to create kubernetes client: failed to load kubeconfig: stat /invalid/path: no such file or directory
```

### API Connection Errors
```bash
$ labrat hub managedclusters
Error: failed to list managed clusters: unable to connect to the server: dial tcp: lookup api.cluster.example.com: no such host
```

### Invalid Status Filter
```bash
$ labrat hub managedclusters --status InvalidStatus
# Returns all clusters (filter is case-sensitive)
```

## Exit Codes

- `0`: Success
- `1`: Error occurred (configuration, connection, or API errors)

## See Also

- [`labrat hub status`](hub-status.md) - Check overall ACM hub health
- [LABRAT Configuration Guide](../configuration.md)
- [ACM ManagedCluster Documentation](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/)

## Notes

- This command requires network access to the ACM hub cluster
- Ensure your kubeconfig has appropriate RBAC permissions to list ManagedCluster resources
- Large clusters (100+ managed clusters) may take several seconds to retrieve
- The command reads from the namespace specified in configuration (default: `open-cluster-management`)
