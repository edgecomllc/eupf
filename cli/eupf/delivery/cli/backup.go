package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func (cli *CLI) newBackupCmd() *cobra.Command {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup management",
	}

	backupCmd.AddCommand(cli.newBackupShowCmd())
	backupCmd.AddCommand(cli.newBackupCreateCmd())
	backupCmd.AddCommand(cli.newBackupRestoreCmd())

	return backupCmd
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

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g. http://127.0.0.1:8081/api/v1)")
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

	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g. http://127.0.0.1:8081/api/v1)")
	return cmd
}
