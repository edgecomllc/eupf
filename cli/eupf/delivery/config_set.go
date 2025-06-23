package delivery

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func (c *CLI) newConfigSetCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Configuration application management",
	}

	loggingCmd := &cobra.Command{
		Use:   "logging",
		Short: "Logging configuration management",
	}
	loggingCmd.AddCommand(c.newSetLoggingLevel())
	loggingCmd.AddCommand(c.newSetLoggingCaller())

	dataPlaneCmd := &cobra.Command{
		Use:   "dataplane",
		Short: "Dataplane configuration management",
	}
	dataPlaneCmd.AddCommand(c.newSetDataPlaneEbpf())
	dataPlaneCmd.AddCommand(c.newSetDataPlaneAddresses())

	pfcpCmd := &cobra.Command{
		Use:   "pfcp",
		Short: "Pfcp configuration management",
	}
	pfcpCmd.AddCommand(c.newSetPFCPN4())
	pfcpCmd.AddCommand(c.newSetPFCPSxa())
	pfcpCmd.AddCommand(c.newSetPFCPSxb())
	pfcpCmd.AddCommand(c.newSetPFCPTimers())

	gtpCmd := &cobra.Command{
		Use:   "gtp",
		Short: "Gtp configuration management",
	}
	gtpCmd.AddCommand(c.newSetGTPPath())

	setCmd.AddCommand(gtpCmd)
	setCmd.AddCommand(pfcpCmd)
	setCmd.AddCommand(dataPlaneCmd)
	setCmd.AddCommand(loggingCmd)

	return setCmd
}

// logging level <level>
func (cli *CLI) newSetLoggingLevel() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "level",
		Short: "Set logging verbosity level",
		Long: `Configure logging verbosity level (debug, info, warn, error, etc.)

Examples:
  eupf config logging level debug
  eupf config logging level info

Overrides configuration parameter: logging_level`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			level := args[0]

			err := cli.usecase.SetLoggingLevel(ctx, level, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to set logging level: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// logging caller
func (cli *CLI) newSetLoggingCaller() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "caller",
		Short: "Toggle function caller information in logs",
		Long: `Enable or disable including source file and line number in log entries.

Examples:
  eupf config logging caller true   # Enable caller info
  eupf config logging caller false  # Disable caller info

Overrides configuration parameter: logging_caller`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			caller := args[0] == "true"

			err := cli.usecase.SetLoggingCaller(ctx, caller, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure logging caller: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// dataplane ebpf --interfaces <ifaces> --xdp-attach <mode>
func (cli *CLI) newSetDataPlaneEbpf() *cobra.Command {
	var baseURL string
	var interfaces []string
	var xdpAttachMode string

	cmd := &cobra.Command{
		Use:   "ebpf",
		Short: "Configure eBPF dataplane interfaces",
		Long: "Bind network interfaces for eBPF dataplane processing\n\n" +
			"Overrides configuration parameters:\n" +
			"  - interface_name\n" +
			"  - xdp_attach_mode",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetDataPlaneEbpf(ctx, interfaces, xdpAttachMode, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure eBPF dataplane: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringArrayVarP(&interfaces, "interfaces", "i", nil, "Network interfaces for eBPF binding (required)")
	cmd.MarkFlagRequired("interfaces")
	cmd.Flags().StringVar(&xdpAttachMode, "xdp-attach", "", "XDP attach mode (generic, native, offload) (required)")
	cmd.MarkFlagRequired("xdp-attach")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// dataplane addresses --n3 <address> --n9 <address>
func (cli *CLI) newSetDataPlaneAddresses() *cobra.Command {
	var baseURL string
	var n3 string
	var n9 string

	cmd := &cobra.Command{
		Use:   "addresses",
		Short: "Configure N3 and N9 interface addresses",
		Long: `Set IP addresses for dataplane N3 and N9 interfaces
Both --n3 and --n9 flags are required.

Overrides configuration parameters:
  - n3_address
  - n9_address`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetDataPlaneAddresses(ctx, n3, n9, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure dataplane addresses: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&n3, "n3", "", "IP address for N3 interface (required)")
	cmd.MarkFlagRequired("n3")
	cmd.Flags().StringVar(&n9, "n9", "", "IP address for N9 interface (required)")
	cmd.MarkFlagRequired("n9")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// pfcp n4 --addr <addr> --node <id> --remote <nodes>
func (cli *CLI) newSetPFCPN4() *cobra.Command {
	var baseURL string
	var pfcpAddress string
	var pfcpNodeId string
	var pfcpRemoteNode []string

	cmd := &cobra.Command{
		Use:   "n4",
		Short: "Configure PFCP N4 connection parameters",
		Long: `Set local and remote node information for PFCP N4 connections
Requires --addr, --node and --remote

Overrides configuration parameters:
  - pfcp_address
  - pfcp_node_id
  - pfcp_node`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetPFCPN4(ctx, pfcpAddress, pfcpNodeId, pfcpRemoteNode, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure PFCP N4: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&pfcpAddress, "addr", "", "Local PFCP address for N4 interface (required)")
	cmd.MarkFlagRequired("addr")
	cmd.Flags().StringVar(&pfcpNodeId, "node", "", "PFCP Node ID identifier (required)")
	cmd.MarkFlagRequired("node")
	cmd.Flags().StringArrayVarP(&pfcpRemoteNode, "remote", "r", nil, "PFCP remote node addresses (required)")
	cmd.MarkFlagRequired("remote")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// pfcp sxa --addr <addr> --node <id> --remote <nodes>
func (cli *CLI) newSetPFCPSxa() *cobra.Command {
	var baseURL string
	var sxaAddress string
	var sxaNodeId string
	var sxaRemoteNode []string

	cmd := &cobra.Command{
		Use:   "sxa",
		Short: "Configure PFCP Sxa connection parameters",
		Long: `Set local and remote node information for PFCP Sxa connections
Requires --addr, --node and --remote

Overrides configuration parameters:
  - sxa_address
  - sxa_node_id
  - sxa_remote_node`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetPFCPSxa(ctx, sxaAddress, sxaNodeId, sxaRemoteNode, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure PFCP Sxa: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&sxaAddress, "addr", "", "Local PFCP address for Sxa interface (required)")
	cmd.MarkFlagRequired("addr")
	cmd.Flags().StringVar(&sxaNodeId, "node", "", "PFCP Node ID for Sxa (required)")
	cmd.MarkFlagRequired("node")
	cmd.Flags().StringArrayVarP(&sxaRemoteNode, "remote", "r", nil, "Sxa remote node addresses (required)")
	cmd.MarkFlagRequired("remote")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// pfcp sxb --addr <addr> --node <id> --remote <nodes>
func (cli *CLI) newSetPFCPSxb() *cobra.Command {
	var baseURL string
	var sxbAddress string
	var sxbNodeId string
	var sxbRemoteNode []string

	cmd := &cobra.Command{
		Use:   "sxb",
		Short: "Configure PFCP Sxb connection parameters",
		Long: `Set local and remote node information for PFCP Sxb connections
Requires --addr, --node and --remote

Overrides configuration parameters:
  - sxb_address
  - sxb_node_id
  - sxb_remote_node`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetPFCPSxb(ctx, sxbAddress, sxbNodeId, sxbRemoteNode, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure PFCP Sxb: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&sxbAddress, "addr", "", "Local PFCP address for Sxb interface (required)")
	cmd.MarkFlagRequired("addr")
	cmd.Flags().StringVar(&sxbNodeId, "node", "", "PFCP Node ID for Sxb (required)")
	cmd.MarkFlagRequired("node")
	cmd.Flags().StringArrayVarP(&sxbRemoteNode, "remote", "r", nil, "Sxb remote node addresses (required)")
	cmd.MarkFlagRequired("remote")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// pfcp timers --assoc-timeout <sec> --heartbeat-timeout <sec>
func (cli *CLI) newSetPFCPTimers() *cobra.Command {
	var baseURL string
	var associationSetupTimeout uint32
	var heartbeatTimeout uint32

	cmd := &cobra.Command{
		Use:   "timers",
		Short: "Configure PFCP session timers",
		Long: `Set timeout values for PFCP association and heartbeat
Requires both --assoc-timeout and --heartbeat-timeout flags.

Overrides configuration parameters:
  - association_setup_timeout
  - heartbeat_timeout`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetPFCPTimers(ctx, associationSetupTimeout, heartbeatTimeout, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure PFCP timers: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().Uint32Var(&associationSetupTimeout, "assoc-timeout", 0, "Association setup timeout in seconds (required)")
	cmd.MarkFlagRequired("assoc-timeout")
	cmd.Flags().Uint32Var(&heartbeatTimeout, "heartbeat-timeout", 0, "Heartbeat timeout in seconds (required)")
	cmd.MarkFlagRequired("heartbeat-timeout")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// gtp path --peers <peers> --echo-interval <sec>
func (cli *CLI) newSetGTPPath() *cobra.Command {
	var baseURL string
	var gtpPeer []string
	var gtpEchoInterval uint32

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Configure GTP path parameters",
		Long: `Set GTP peers and echo interval for path management
Requires at least one GTP peer and echo interval ≥1 second.
Examples:
  gtp path --peers 10.0.0.1 --echo-interval 30
  gtp path -p 192.168.1.1 -p 192.168.1.2 --echo-interval 60

Overrides configuration parameters:
  - gtp_peer
  - gtp_echo_interval`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.SetGTPPath(ctx, gtpPeer, gtpEchoInterval, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure GTP path: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringArrayVarP(&gtpPeer, "peers", "p", nil, "GTP peer addresses (required)")
	cmd.MarkFlagRequired("peers")
	cmd.Flags().Uint32Var(&gtpEchoInterval, "echo-interval", 1, "GTP echo interval in seconds (required)")
	cmd.MarkFlagRequired("echo-interval")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address")

	return cmd
}
