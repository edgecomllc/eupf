package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/edgecomllc/eupf/cli/domain"
	"github.com/spf13/cobra"
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

	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup management",
	}

	backupCmd.AddCommand(cli.newBackupShowCmd())
	backupCmd.AddCommand(cli.newBackupCreateCmd())
	backupCmd.AddCommand(cli.newBackupRestoreCmd())

	root.AddCommand(traceCmd)
	root.AddCommand(backupCmd)
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

// backup show
func (cli *CLI) newBackupShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show backups files",
		Long:  "Retrieve list of backups.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			backups, err := cli.usecase.BackupList(ctx)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "failed to get backup list: %v\n", err)
				os.Exit(1)
			}

			if len(backups) == 0 {
				fmt.Println("No saved backups.")
				return
			}

			sort.Slice(backups, func(i, j int) bool {
				return !backups[i].Timestamp.Before(backups[j].Timestamp)
			})

			for i, b := range backups {
				fmt.Printf("#%d: %s (%s)\n", i+1, b.Name, b.Timestamp.Format("02.01.06 15:04:05 MST"))
			}
		},
	}

	return cmd
}

// backup create
func (cli *CLI) newBackupCreateCmd() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create backup",
		Long:  "Create backup configuration from current settings. Use --baseurl to override the eupf endpoint .",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cli.usecase.BackupCreate(ctx, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create backup: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g. http://localhost:8081)")
	return cmd
}

// backup restore <id of backup>
func (cli *CLI) newBackupRestoreCmd() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "restore <id of backup>",
		Short: "Restore backup",
		Long: "Restore backup configuration from given id. Use --baseurl to override the eupf endpoint.\n\n" +
			"ID of backup can be found from 'backup show' command. Example: 'eupf backup restore 1745307978'",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			backupID := args[0]

			err := cli.usecase.BackupRestore(ctx, baseURL, backupID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create backup: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g. http://localhost:8081)")
	return cmd
}
