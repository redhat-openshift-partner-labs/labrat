package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redhat-openshift-partner-labs/labrat/internal/config"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/kube"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/spoke"
	"github.com/spf13/cobra"
)

// version of the tool (can be set via ldflags during build)
var version = "0.1.0" //nolint:unused // will be used in future version command

func main() {
	rootCmd := &cobra.Command{
		Use:   "labrat",
		Short: "Lab Administration, Bootstrapping, and Resource Automation Toolkit",
		Long: `LABRAT is the primary CLI utility for the OpenShift Partner Labs offering.
It provides a centralized interface for managing the ACM Hub and partner spoke clusters.`,
	}

	// Persistent Flags
	rootCmd.PersistentFlags().StringP("config", "c", "$HOME/.labrat/config.yaml", "path to labrat config")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable debug logging")

	// --- HUB COMMAND ---
	hubCmd := &cobra.Command{
		Use:   "hub",
		Short: "Interact with the primary ACM management cluster",
	}
	hubStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check health of the ACM hub",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("🔍 Checking ACM Hub status...")
			// Logic for OpenShift API calls would go here
		},
	}

	hubManagedClustersCmd := &cobra.Command{
		Use:   "managedclusters",
		Short: "List ACM managed clusters",
		Long:  `List all managed clusters from the ACM hub with status information.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// 1. Get flags
			configPath, _ := cmd.Flags().GetString("config")
			outputFormat, _ := cmd.Flags().GetString("output")
			statusFilter, _ := cmd.Flags().GetString("status")
			wide, _ := cmd.Flags().GetBool("wide")

			// 2. Load config (expand path to support both $HOME and ~)
			cfg, err := config.Load(config.ExpandPath(configPath))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// 3. Create Kubernetes client
			kubeClient, err := kube.NewClient(cfg.GetHubKubeconfig(), cfg.Hub.Context)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// 4. Create output writer
			output := hub.NewOutputWriter(hub.OutputFormat(outputFormat), os.Stdout)

			// 5. If --wide flag is set, use combined cluster view
			ctx := context.Background()
			if wide {
				// Create both ManagedCluster and ClusterDeployment clients
				mcClient := hub.NewManagedClusterClient(kubeClient.GetDynamicClient())
				cdClient := hub.NewClusterDeploymentClient(kubeClient.GetDynamicClient())
				combinedClient := hub.NewCombinedClusterClient(mcClient, cdClient)

				// List combined clusters
				combined, err := combinedClient.ListCombined(ctx)
				if err != nil {
					return fmt.Errorf("failed to list combined clusters: %w", err)
				}

				// Apply filter if specified (filter on Status field)
				if statusFilter != "" {
					filtered := make([]hub.CombinedClusterInfo, 0)
					for _, cluster := range combined {
						if string(cluster.Status) == statusFilter {
							filtered = append(filtered, cluster)
						}
					}
					combined = filtered
				}

				// Output combined results
				if err := output.WriteCombined(combined, true); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
			} else {
				// Use standard ManagedCluster view
				mcClient := hub.NewManagedClusterClient(kubeClient.GetDynamicClient())

				// List clusters
				clusters, err := mcClient.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list managed clusters: %w", err)
				}

				// Apply filter if specified
				if statusFilter != "" {
					filter := hub.ManagedClusterFilter{
						Status: hub.ClusterStatus(statusFilter),
					}
					clusters = mcClient.Filter(clusters, filter)
				}

				// Output results
				if err := output.Write(clusters); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
			}

			return nil
		},
	}

	hubManagedClustersCmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	hubManagedClustersCmd.Flags().String("status", "", "Filter by status (Ready|NotReady|Unknown)")
	hubManagedClustersCmd.Flags().Bool("wide", false, "Show additional cluster details from ClusterDeployment")

	hubCmd.AddCommand(hubStatusCmd, hubManagedClustersCmd)

	// --- SPOKE COMMAND ---
	spokeCmd := &cobra.Command{
		Use:   "spoke",
		Short: "Manage individual partner-requested clusters",
	}
	spokeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Provision a new partner cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			requestID, err := cmd.Flags().GetString("request-id")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting request-id: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("🚀 Initiating bootstrap for request: %s\n", requestID)
		},
	}
	spokeCreateCmd.Flags().String("request-id", "", "ID of the partner request (Required)")
	if err := spokeCreateCmd.MarkFlagRequired("request-id"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag required: %v\n", err)
		os.Exit(1)
	}

	spokeKubeconfigCmd := &cobra.Command{
		Use:   "kubeconfig <cluster-name>",
		Short: "Extract admin kubeconfig for a spoke cluster",
		Long: `Extract the admin kubeconfig from a spoke cluster's ClusterDeployment secret.

This command retrieves the admin kubeconfig which has full cluster-admin privileges.
Use with caution and store securely.

Examples:
  # Print kubeconfig to stdout
  labrat spoke kubeconfig my-cluster

  # Save kubeconfig to file
  labrat spoke kubeconfig my-cluster -o /tmp/my-cluster.kubeconfig

  # Use the kubeconfig with kubectl
  labrat spoke kubeconfig my-cluster -o /tmp/kubeconfig
  kubectl --kubeconfig /tmp/kubeconfig get nodes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName := args[0]
			configPath, _ := cmd.Flags().GetString("config")
			outputPath, _ := cmd.Flags().GetString("output")

			// Load config
			cfg, err := config.Load(config.ExpandPath(configPath))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create Kubernetes client
			kubeClient, err := kube.NewClient(cfg.GetHubKubeconfig(), cfg.Hub.Context)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// Create kubeconfig extractor
			extractor := spoke.NewKubeconfigExtractor(
				kubeClient.GetDynamicClient(),
				kubeClient.GetCoreClient().CoreV1(),
			)

			ctx := context.Background()

			// Display security warning
			fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: This is an admin kubeconfig with full cluster-admin privileges!\n")
			fmt.Fprintf(os.Stderr, "    Please store it securely and restrict access appropriately.\n\n")

			if outputPath != "" {
				// Extract to file
				if err := extractor.ExtractToFile(ctx, clusterName, outputPath); err != nil {
					return fmt.Errorf("failed to extract kubeconfig: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✓ Kubeconfig saved to: %s\n", outputPath)
				fmt.Fprintf(os.Stderr, "  File permissions set to 0600 (owner read/write only)\n\n")
				fmt.Fprintf(os.Stderr, "You can now use it with kubectl:\n")
				fmt.Fprintf(os.Stderr, "  kubectl --kubeconfig %s get nodes\n", outputPath)
			} else {
				// Extract to stdout
				kubeconfig, err := extractor.Extract(ctx, clusterName)
				if err != nil {
					return fmt.Errorf("failed to extract kubeconfig: %w", err)
				}
				fmt.Print(string(kubeconfig))
			}

			return nil
		},
	}
	spokeKubeconfigCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")

	spokeCmd.AddCommand(spokeCreateCmd, spokeKubeconfigCmd)

	// --- BOOTSTRAP COMMAND ---
	bootstrapCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Initialize new lab environments",
	}
	bootstrapInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize local labrat configuration",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("⚙️ Initializing LABRAT environment...")
		},
	}
	bootstrapCmd.AddCommand(bootstrapInitCmd)

	// Add all top-level commands to root
	rootCmd.AddCommand(hubCmd, spokeCmd, bootstrapCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
