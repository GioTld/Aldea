package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/GioTld/aldea/internal/runtime"
	"github.com/GioTld/aldea/internal/scheduler"
	"github.com/spf13/cobra"
)

var globalRuntimeEngine = runtime.NewEngine()

func NewComputeCmd() *cobra.Command {
	computeCmd := &cobra.Command{
		Use:   "compute",
		Short: "Manage distributed compute workloads across cluster nodes",
		Long:  "Deploy, list, and stop OCI compute workloads running on compute-enabled Linux nodes.",
	}

	var autoConfirm bool

	deployCmd := &cobra.Command{
		Use:   "deploy <manifest.yaml>",
		Short: "Deploy an OCI compute workload from a manifest YAML file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("reading manifest file: %w", err)
			}

			m, err := scheduler.ParseManifestYAML(data)
			if err != nil {
				return fmt.Errorf("invalid workload manifest: %w", err)
			}

			// Surface explicit user-facing security warning without requirement codes (RNF-19)
			if !autoConfirm {
				cmd.Printf("⚠️ SECURITY WARNING: Compute workloads execute in plaintext RAM on node host machines. The owner of the assigned node can inspect code and in-memory data during execution.\nDo you wish to proceed? [y/N]: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToLower(input))
				if input != "y" && input != "yes" {
					cmd.Println("Aborted deployment.")
					return nil
				}
			}

			// Run scheduler placement
			sched := scheduler.NewScheduler()
			candidates := []scheduler.NodeComputeCapability{
				{
					NodeID:            "local-node",
					OS:                "linux",
					ComputeEnabled:    true,
					IsHealthy:         true,
					TotalCPUCores:     4.0,
					AllocatedCPUCores: 0,
					TotalMemoryMB:     4096,
					AllocatedMemoryMB: 0,
				},
			}

			selectedNode, err := sched.SelectNode(*m, candidates)
			if err != nil {
				return fmt.Errorf("scheduling workload: %w", err)
			}

			cmd.Printf("[+] Scheduling workload '%s' (%s) to node %s...\n", m.Name, m.WorkloadID, selectedNode.NodeID)

			// Start container in runtime engine
			status, err := globalRuntimeEngine.StartWorkload(context.Background(), *m)
			if err != nil {
				return fmt.Errorf("starting compute workload: %w", err)
			}

			cmd.Printf("[✓] Workload deployed successfully. ID: %s | State: %s | IP: %s\n", status.WorkloadID, status.State, status.IPAddress)
			return nil
		},
	}
	deployCmd.Flags().BoolVarP(&autoConfirm, "yes", "y", false, "Automatically confirm security warning and proceed")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all active compute workloads",
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := globalRuntimeEngine.ListWorkloads(context.Background())
			if err != nil {
				return fmt.Errorf("listing workloads: %w", err)
			}

			cmd.Println("WORKLOAD ID\tNAME\tIMAGE\tSTATE\tIP ADDRESS")
			for _, w := range list {
				cmd.Printf("%s\t%s\t%s\t%s\t%s\n", w.WorkloadID, w.Name, w.Image, w.State, w.IPAddress)
			}
			return nil
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop <workloadID>",
		Short: "Stop a running compute workload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workloadID := args[0]
			err := globalRuntimeEngine.StopWorkload(context.Background(), workloadID)
			if err != nil {
				return fmt.Errorf("stopping workload: %w", err)
			}
			cmd.Printf("[✓] Workload %s stopped successfully.\n", workloadID)
			return nil
		},
	}

	computeCmd.AddCommand(deployCmd, listCmd, stopCmd)
	return computeCmd
}
