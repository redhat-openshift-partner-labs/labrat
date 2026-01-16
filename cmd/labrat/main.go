package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redhat-openshift-partner-labs/labrat/internal/config"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/hub"
	"github.com/redhat-openshift-partner-labs/labrat/pkg/kube"
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
	spokeCmd.AddCommand(spokeCreateCmd)

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
