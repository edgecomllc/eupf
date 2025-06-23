package delivery

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func (c *CLI) newConfigShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration settings",
	}

	loggingCmd := &cobra.Command{
		Use:   "logging",
		Short: "Display logging configuration",
	}
	loggingCmd.AddCommand(c.newShowLoggingLevel())
	loggingCmd.AddCommand(c.newShowLoggingCaller())

	dataPlaneCmd := &cobra.Command{
		Use:   "dataplane",
		Short: "Display dataplane configuration",
	}
	dataPlaneCmd.AddCommand(c.newShowDataPlaneEbpf())
	dataPlaneCmd.AddCommand(c.newShowDataPlaneAddresses())

	pfcpCmd := &cobra.Command{
		Use:   "pfcp",
		Short: "Display PFCP configuration",
	}
	pfcpCmd.AddCommand(c.newShowPFCPN4())
	pfcpCmd.AddCommand(c.newShowPFCPSxa())
	pfcpCmd.AddCommand(c.newShowPFCPSxb())
	pfcpCmd.AddCommand(c.newShowPFCPTimers())

	gtpCmd := &cobra.Command{
		Use:   "gtp",
		Short: "Display GTP configuration",
	}
	gtpCmd.AddCommand(c.newShowGTPPath())

	showCmd.AddCommand(gtpCmd)
	showCmd.AddCommand(pfcpCmd)
	showCmd.AddCommand(dataPlaneCmd)
	showCmd.AddCommand(loggingCmd)

	return showCmd
}

// logging level
func (cli *CLI) newShowLoggingLevel() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "level",
		Short: "Show current logging verbosity level",
		Long:  `Display current logging verbosity level (debug, info, warn, error, etc.)`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Level: %s\n", cfg.LoggingLevel)
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// logging caller
func (cli *CLI) newShowLoggingCaller() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "caller",
		Short: "Show current caller information setting",
		Long:  `Display whether source file and line numbers are included in log entries`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Caller Info Enabled: %t\n", cfg.LoggingCaller)
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// dataplane ebpf
func (cli *CLI) newShowDataPlaneEbpf() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "ebpf",
		Short: "Show current eBPF dataplane configuration",
		Long:  "Display current network interfaces bound for eBPF processing and XDP attach mode",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("XDP Attach Mode: %s\n", cfg.XDPAttachMode)
			fmt.Println("Interfaces:")
			for i, iface := range cfg.InterfaceName {
				fmt.Printf("  %d. %s\n", i+1, iface)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// dataplane addresses
func (cli *CLI) newShowDataPlaneAddresses() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "addresses",
		Short: "Show current N3/N9 interface addresses",
		Long:  "Display current IP addresses configured for dataplane N3 and N9 interfaces",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("N3 Address: %s\n", cfg.N3Address)
			fmt.Printf("N9 Address: %s\n", cfg.N9Address)
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// pfcp n4
func (cli *CLI) newShowPFCPN4() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "n4",
		Short: "Show current PFCP N4 configuration",
		Long:  "Display local and remote node information for PFCP N4 connections",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Local Address: %s\n", cfg.PfcpAddress)
			fmt.Printf("Node ID: %s\n", cfg.PfcpNodeId)
			fmt.Println("Remote Nodes:")
			for i, node := range cfg.PfcpRemoteNode {
				fmt.Printf("  %d. %s\n", i+1, node)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// pfcp sxa
func (cli *CLI) newShowPFCPSxa() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "sxa",
		Short: "Show current PFCP Sxa configuration",
		Long:  "Display local and remote node information for PFCP Sxa connections",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Local Address: %s\n", cfg.SxaLocalAddress)
			fmt.Printf("Node ID: %s\n", cfg.SxaLocalNodeId)
			fmt.Println("Remote Nodes:")
			for i, node := range cfg.SxaRemoteNode {
				fmt.Printf("  %d. %s\n", i+1, node)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// pfcp sxb
func (cli *CLI) newShowPFCPSxb() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "sxb",
		Short: "Show current PFCP Sxb configuration",
		Long:  "Display local and remote node information for PFCP Sxb connections",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Local Address: %s\n", cfg.SxbLocalAddress)
			fmt.Printf("Node ID: %s\n", cfg.SxbLocalNodeId)
			fmt.Println("Remote Nodes:")
			for i, node := range cfg.SxbRemoteNode {
				fmt.Printf("  %d. %s\n", i+1, node)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// pfcp timers
func (cli *CLI) newShowPFCPTimers() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "timers",
		Short: "Show current PFCP timer configuration",
		Long:  "Display timeout values for PFCP association and heartbeat",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Association Setup Timeout: %d seconds\n", cfg.AssociationSetupTimeout)
			fmt.Printf("Heartbeat Timeout: %d seconds\n", cfg.HeartbeatTimeout)
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}

// gtp path
func (cli *CLI) newShowGTPPath() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show current GTP path configuration",
		Long:  "Display GTP peers and echo interval for path management",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cfg, err := cli.usecase.GetConfig(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error retrieving configuration: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Echo Interval: %d seconds\n", cfg.GtpEchoInterval)
			fmt.Println("GTP Peers:")
			for i, peer := range cfg.GtpPeer {
				fmt.Printf("  %d. %s\n", i+1, peer)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")
	return cmd
}
