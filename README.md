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

# Build the binary
task build

# Run the tool locally
task run -- hub status

# Install the binary to your $GOPATH/bin
task install
```

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