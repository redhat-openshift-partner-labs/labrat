# LABRAT 🐀

**L**ab **A**dministration, **B**ootstrapping, and **R**esource **A**utomation **T**oolkit

> **Mission Statement:** To empower the OpenShift Partner Labs ecosystem by providing a unified, automated interface for the seamless orchestration of cluster lifecycles. LABRAT eliminates the complexity of manual administration, ensuring that from the ACM Hub to the edge spokes, every resource is bootstrapped with precision and maintained with ease.

## 🚀 Overview

LABRAT is the specialized command-line utility for the **OpenShift Partner Labs** offering. It serves as the primary interface for managing the OpenShift Hub (running Advanced Cluster Management/ACM) and the fleet of partner-requested "spoke" clusters.

### Core Capabilities

* **ACM Hub Management**: Direct interaction with the management cluster to monitor health and policies.
* **Automated Bootstrapping**: Rapid deployment of partner lab environments with standardized configurations.
* **Resource Lifecycle**: Automated maintenance, patching, and decommissioning of ephemeral lab resources.

## ⌨️ CLI Command Structure

LABRAT follows a standard POSIX-style command hierarchy:

```text
Usage:
  labrat [command] [subcommand] [flags]

Commands:
  hub        Interact with the primary ACM management cluster (status, logs, etc.)
  spoke      Manage individual partner clusters (create, delete, list)
  bootstrap  Initialize local environments or provision new lab templates
  sync       Reconcile configuration state between Hub and Spokes
  status     Global health overview of the lab ecosystem

Global Flags:
  -c, --config      Path to labrat config (default: ~/.labrat/config.yaml)
  --kubeconfig      Path to kubeconfig (default: ~/.kube/config)
  -v, --verbose     Enable debug logging
```

## 📖 Commands

### Hub Commands

#### `labrat hub managedclusters`

List all ACM managed clusters from the hub with status information.

**Usage**:
```bash
labrat hub managedclusters [flags]
```

**Flags**:
- `--output, -o`: Output format (table|json), default: table
- `--status`: Filter by status (Ready|NotReady|Unknown), optional
- `--config, -c`: Path to labrat config (default: ~/.labrat/config.yaml)
- `--verbose, -v`: Enable debug logging

**Examples**:

```bash
# List all managed clusters in table format
labrat hub managedclusters

# Output as JSON
labrat hub managedclusters --output json

# Filter by status
labrat hub managedclusters --status Ready
labrat hub managedclusters --status NotReady

# Use custom config
labrat hub managedclusters --config ./my-config.yaml
```

**Example Output** (table format):
```
NAME                STATUS      AVAILABLE
cluster-east-1      Ready       True
cluster-west-1      NotReady    False
cluster-central     Unknown     Unknown
```

**Prerequisites**:
- Access to an ACM hub cluster
- Valid kubeconfig configured in `~/.labrat/config.yaml`
- Kubernetes permissions to list ManagedCluster resources

**Status Derivation**:
The command derives cluster status using the following priority:
1. Unreachable taint present → NotReady
2. ManagedClusterConditionAvailable=True → Ready
3. ManagedClusterConditionAvailable=False → NotReady
4. ManagedClusterConditionAvailable=Unknown → Unknown
5. No conditions → Unknown

## 🛠 Development & Build

This project uses [Taskfile](https://taskfile.dev) for task automation.

### Prerequisites

* Go 1.25+
* [Task](https://taskfile.dev/install/) installed on your system
* Access to an OpenShift/ACM environment

### Getting Started

```bash
# Initialize Go modules and dependencies
task init

# Set up configuration (choose one option):

# Option 1: Use the development config in the project directory
# No setup needed - use --config flag with commands

# Option 2: Copy to standard location
mkdir -p ~/.labrat
cp config.yaml ~/.labrat/config.yaml

# Option 3: Symlink for automatic updates during development
mkdir -p ~/.labrat
ln -s $(pwd)/config.yaml ~/.labrat/config.yaml

# Build the binary
task build

# Run the tool locally (using project config)
./bin/labrat hub managedclusters --config config.yaml

# Or run with standard config location (~/.labrat/config.yaml)
./bin/labrat hub managedclusters

# Install the binary to your $GOPATH/bin
task install
```

### Configuration

LABRAT requires a configuration file to connect to your ACM hub cluster. A development-ready `config.yaml` is included in the project root.

**Default config location**: `~/.labrat/config.yaml`

**Development options**:
1. Use project config: `labrat [command] --config config.yaml`
2. Copy to home directory: `cp config.yaml ~/.labrat/config.yaml`
3. Symlink for auto-sync: `ln -s $(pwd)/config.yaml ~/.labrat/config.yaml`

**Required configuration**:
- `hub.kubeconfig`: Path to kubeconfig for ACM hub cluster
- `hub.namespace`: ACM namespace (default: `open-cluster-management`)

See `config.yaml` for full configuration options and documentation.

## 📂 Project Structure

Following the standard Go project layout:

* `cmd/labrat/`: Main entry point and CLI command definitions.
* `pkg/`: Public library logic for Hub and Spoke management.
* `internal/`: Private utility code (configuration parsing, internal helpers).
* `bin/`: Compiled binaries (ignored by git).
* `Taskfile.yaml`: Project automation and build tasks.

## 🏗 Architecture

LABRAT acts as the orchestration layer between:

1. **The Hub**: The central authority running Red Hat ACM.
2. **The Spokes**: The distributed OpenShift clusters provisioned for partners across various cloud and on-premise providers.

---
*Maintained by the OpenShift Partner Labs Team.*