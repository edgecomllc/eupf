package cli

import (
	"context"
	"fmt"
	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type CLI struct {
	Root    *cobra.Command
	usecase domain.EupfUseCase
}

func NewCLI(usecase domain.EupfUseCase) *CLI {
	root := &cobra.Command{
		Use:   "eupf",
		Short: "eUPF CLI tool",
	}

	cli := &CLI{
		Root:    root,
		usecase: usecase,
	}

	traceCmd := &cobra.Command{
		Use:   "trace",
		Short: "Trace management",
	}

	traceCmd.AddCommand(cli.newTraceShowCmd())
	traceCmd.AddCommand(cli.newTraceSetCmd())
	traceCmd.AddCommand(cli.newTraceStopCmd())

	root.AddCommand(traceCmd)
	root.AddCommand(cli.newCompletion())

	return cli
}

func (cli *CLI) Execute() error {
	return cli.Root.Execute()
}

// trace show
func (cli *CLI) newTraceShowCmd() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current trace records",
		Long:  "Retrieve active trace records from the eupf API. Optionally use --baseurl to override target.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			traces, err := cli.usecase.TraceList(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to list traces: %v\n", err)
				os.Exit(1)
			}

			if len(traces) == 0 {
				fmt.Println("No active traces.")
				return
			}

			for _, t := range traces {
				fmt.Println(t)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g. http://localhost:8081)")
	return cmd
}

func (cli *CLI) newCompletion() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			err := cli.Root.GenBashCompletion(os.Stdout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "completion generation failed: %v\n", err)
				os.Exit(1)
			}
		},
	}
}

// trace set --imsi x --msisdn y
func (cli *CLI) newTraceSetCmd() *cobra.Command {
	var imsi string
	var msisdn string
	var baseURL string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Start trace for IMSI and/or MSISDN",
		Long:  "Start subscriber trace session. Use --imsi and/or --msisdn to filter, and --baseurl to override the eupf endpoint.",
		Run: func(cmd *cobra.Command, args []string) {
			if imsi == "" && msisdn == "" {
				fmt.Fprintln(os.Stderr, "error: at least one of --imsi or --msisdn must be provided")
				os.Exit(1)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.StartTrace(ctx, &imsi, &msisdn, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to start trace: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Trace started (imsi=%s, msisdn=%s)\n", imsi, msisdn)
		},
	}

	cmd.Flags().StringVar(&imsi, "imsi", "", "IMSI to trace")
	cmd.Flags().StringVar(&msisdn, "msisdn", "", "MSISDN to trace")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API")

	return cmd
}

// trace stop --imsi x --msisdn y
func (cli *CLI) newTraceStopCmd() *cobra.Command {
	var imsi string
	var msisdn string
	var baseURL string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop trace for IMSI and/or MSISDN",
		Long:  "Stop subscriber trace session. Use --imsi and/or --msisdn to match, and --baseurl to override the eupf endpoint.",
		Run: func(cmd *cobra.Command, args []string) {
			if imsi == "" && msisdn == "" {
				fmt.Fprintln(os.Stderr, "error: at least one of --imsi or --msisdn must be provided")
				os.Exit(1)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.StopTrace(ctx, &imsi, &msisdn, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to stop trace: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Trace stopped (imsi=%s, msisdn=%s)\n", imsi, msisdn)
		},
	}

	cmd.Flags().StringVar(&imsi, "imsi", "", "IMSI to stop trace")
	cmd.Flags().StringVar(&msisdn, "msisdn", "", "MSISDN to stop trace")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API")

	return cmd
}
