# Command: `labrat spoke kubeconfig`

## Synopsis

Extract the admin kubeconfig for a spoke cluster from the ACM hub cluster.

```bash
labrat spoke kubeconfig <cluster-name> [flags]
```

## Description

The `kubeconfig` subcommand retrieves the admin kubeconfig for a managed spoke cluster from the hub cluster. The kubeconfig is stored in a Kubernetes secret within the ClusterDeployment resource on the hub and contains credentials with cluster-admin privileges for the spoke cluster.

This command is useful for:
- Obtaining direct access to a spoke cluster for troubleshooting
- Setting up local kubectl access to managed clusters
- Automating spoke cluster operations from external scripts
- Debugging connectivity or authentication issues

**Security Note**: The extracted kubeconfig contains cluster-admin credentials. Handle it securely and avoid committing it to version control.

## Arguments

### `<cluster-name>`
**Required**: Yes
**Type**: String

The name of the managed cluster for which to extract the kubeconfig. This must match the name of a ManagedCluster resource on the hub.

## Flags

### `--output, -o`
**Type**: String
**Optional**: Yes
**Default**: Prints to stdout

Specifies the output file path where the kubeconfig should be written. If not provided, the kubeconfig is printed to stdout.

**Example**:
```bash
# Write to file
labrat spoke kubeconfig my-cluster --output /path/to/kubeconfig

# Short form
labrat spoke kubeconfig my-cluster -o ~/.kube/my-cluster-config

# Print to stdout (default)
labrat spoke kubeconfig my-cluster
```

### Global Flags

Inherited from parent commands:

- `--config, -c`: Path to labrat config file (default: `$HOME/.labrat/config.yaml`)
- `--verbose, -v`: Enable debug logging

## Output

The command outputs a complete kubeconfig file in YAML format containing:

- Cluster API server endpoint
- Certificate authority data
- Client certificate and key (cluster-admin credentials)
- Context configuration

**Example Output** (stdout):
```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://api.my-cluster.example.com:6443
  name: my-cluster
contexts:
- context:
    cluster: my-cluster
    user: admin
  name: admin
current-context: admin
users:
- name: admin
  user:
    client-certificate-data: LS0tLS1CRUdJTi...
    client-key-data: LS0tLS1CRUdJTi...
```

## How It Works

1. **Lookup ClusterDeployment**: The command searches for a ClusterDeployment resource with the specified cluster name in the namespace matching the cluster name
2. **Extract Secret Reference**: Retrieves the admin kubeconfig secret reference from the ClusterDeployment metadata
3. **Fetch Secret**: Reads the Kubernetes secret containing the kubeconfig data
4. **Output**: Writes the kubeconfig to the specified file or stdout

## Examples

### Extract and Save to File
```bash
labrat spoke kubeconfig my-cluster --output ~/.kube/my-cluster-config
```

Output:
```
Successfully extracted kubeconfig for cluster 'my-cluster' to /home/user/.kube/my-cluster-config
```

### Use Directly with kubectl
```bash
# Extract to file
labrat spoke kubeconfig prod-cluster -o /tmp/prod-kubeconfig

# Use with kubectl
kubectl --kubeconfig /tmp/prod-kubeconfig get nodes
```

### Pipe to kubectl
```bash
# Extract to stdout and use with kubectl
labrat spoke kubeconfig dev-cluster | kubectl --kubeconfig /dev/stdin get pods -A
```

### Set as Default kubeconfig
```bash
# Extract and merge with existing kubeconfig
labrat spoke kubeconfig my-cluster -o ~/.kube/config-my-cluster

# Export for current session
export KUBECONFIG=~/.kube/config-my-cluster

# Or merge into default kubeconfig
KUBECONFIG=~/.kube/config:~/.kube/config-my-cluster kubectl config view --flatten > ~/.kube/config-merged
mv ~/.kube/config-merged ~/.kube/config
```

### Use with Custom Hub Config
```bash
labrat spoke kubeconfig my-cluster \
  --config /path/to/custom/config.yaml \
  --output /tmp/my-cluster-kubeconfig
```

### Extract for Multiple Clusters
```bash
# Extract kubeconfig for all ready clusters
for cluster in $(labrat hub managedclusters --status Ready -o json | jq -r '.[].name'); do
  labrat spoke kubeconfig "$cluster" -o ~/.kube/spoke-"$cluster"
done
```

### Enable Verbose Logging
```bash
labrat spoke kubeconfig my-cluster --verbose -o /tmp/kubeconfig
```

## Configuration

This command requires a valid LABRAT configuration file with hub cluster settings.

**Required Configuration** (`~/.labrat/config.yaml`):
```yaml
hub:
  kubeconfig: /path/to/hub/kubeconfig
  context: hub-cluster-context  # Optional, uses current context if not specified
```

## Error Handling

### Cluster Not Found
```bash
$ labrat spoke kubeconfig nonexistent-cluster
Error: ClusterDeployment not found for cluster 'nonexistent-cluster': ensure the cluster exists and has a ClusterDeployment resource
```

### Secret Not Found
```bash
$ labrat spoke kubeconfig my-cluster
Error: admin kubeconfig secret not found for cluster 'my-cluster': the ClusterDeployment may not have completed provisioning
```

### No Secret Reference in ClusterDeployment
```bash
$ labrat spoke kubeconfig my-cluster
Error: admin kubeconfig secret reference not found in ClusterDeployment 'my-cluster': the cluster may not be fully provisioned
```

### Hub Connection Errors
```bash
$ labrat spoke kubeconfig my-cluster
Error: failed to connect to hub cluster: unable to connect to the server: dial tcp: lookup api.hub.example.com: no such host
```

### Invalid Output Path
```bash
$ labrat spoke kubeconfig my-cluster -o /invalid/directory/kubeconfig
Error: failed to write kubeconfig to file: open /invalid/directory/kubeconfig: no such file or directory
```

### Permission Denied
```bash
$ labrat spoke kubeconfig my-cluster
Error: failed to get ClusterDeployment: clusterdeployments.hive.openshift.io "my-cluster" is forbidden: User "system:serviceaccount:default:labrat" cannot get resource "clusterdeployments" in API group "hive.openshift.io" in the namespace "my-cluster"
```

## Exit Codes

- `0`: Success - kubeconfig extracted successfully
- `1`: Error occurred (cluster not found, permission denied, connection errors, etc.)

## Security Considerations

### Credentials

The extracted kubeconfig contains:
- **Cluster-admin privileges**: Full administrative access to the spoke cluster
- **Client certificates**: Long-lived credentials (typically valid for years)
- **Cluster CA**: Certificate authority for validating the API server

**Best Practices**:
- Store kubeconfig files with restrictive permissions (0600)
- Never commit kubeconfig files to version control
- Rotate credentials regularly using ClusterDeployment operations
- Use separate kubeconfigs for different purposes (automation vs manual access)

### File Permissions

When saving to a file, ensure proper permissions:

```bash
# Extract with secure permissions
labrat spoke kubeconfig my-cluster -o ~/.kube/my-cluster
chmod 600 ~/.kube/my-cluster
```

### Hub RBAC Requirements

The user running this command needs the following permissions on the hub cluster:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: labrat-spoke-kubeconfig-extractor
rules:
- apiGroups: ["hive.openshift.io"]
  resources: ["clusterdeployments"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
```

**Note**: Secret access is limited to the admin kubeconfig secret in the ClusterDeployment namespace.

## Troubleshooting

### Cluster Exists but No ClusterDeployment

Some managed clusters may not have a ClusterDeployment resource if they were:
- Imported (not provisioned by Hive)
- Created through different mechanisms
- Running on bare metal or non-cloud platforms

**Solution**: Use alternative methods to obtain kubeconfig for imported clusters.

### Kubeconfig Not Yet Available

If a cluster was recently created, the kubeconfig secret may not be populated yet:

```bash
# Check cluster provisioning status
labrat hub managedclusters --status Unknown

# Wait for cluster to be fully provisioned
labrat hub managedclusters --status Ready
```

### Invalid or Expired Kubeconfig

If the extracted kubeconfig doesn't work:

1. Verify the spoke cluster is healthy:
   ```bash
   labrat hub managedclusters | grep my-cluster
   ```

2. Check the ClusterDeployment status on the hub:
   ```bash
   kubectl --kubeconfig $(cat ~/.labrat/config.yaml | grep kubeconfig | awk '{print $2}') \
     get clusterdeployment my-cluster -n my-cluster -o yaml
   ```

3. Consider re-provisioning or rotating credentials through ACM/Hive

## Related Commands

- [`labrat hub managedclusters`](hub-managedclusters.md) - List managed clusters to find cluster names
- [`labrat spoke get`](spoke-get.md) - Get detailed information about a spoke cluster (planned)

## See Also

- [LABRAT Configuration Guide](../configuration.md)
- [Hive ClusterDeployment Documentation](https://github.com/openshift/hive/blob/master/docs/using-hive.md)
- [ACM Documentation](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/)
- [kubectl Configuration Best Practices](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)

## Notes

- This command requires network access to the ACM hub cluster
- The kubeconfig extraction is read-only; it does not modify cluster resources
- Kubeconfig secrets are typically named `<cluster-name>-admin-kubeconfig`
- The command works with Hive-provisioned clusters (AWS, Azure, GCP, etc.)
- For imported clusters without ClusterDeployment, manual kubeconfig extraction may be required
