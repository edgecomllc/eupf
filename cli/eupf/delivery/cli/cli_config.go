package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (cli *CLI) newCliConfigCmd() *cobra.Command {
	cliConfigCmd := &cobra.Command{
		Use:   "cli-config",
		Short: "Configuration management",
	}

	cliConfigCmd.AddCommand(cli.newConfigSetBaseURLCmd())
	cliConfigCmd.AddCommand(cli.newConfigShowBaseURLCmd())

	return cliConfigCmd
}

// config set-baseurl --url <string>
func (cli *CLI) newConfigSetBaseURLCmd() *cobra.Command {
	var newBaseURL string

	cmd := &cobra.Command{
		Use:   "set-baseurl",
		Short: "Set new base URL in config",
		Long: `Set a new base URL to be used as the default endpoint for eupf API.
This overrides the "eupf_addr" value in the config-cli.yaml file.
Example:
  eupf config set-baseurl --url http://localhost:8081/api/v1`,
		Run: func(cmd *cobra.Command, args []string) {
			if newBaseURL == "" {
				fmt.Fprintln(os.Stderr, "error: --url must be provided")
				os.Exit(1)
			}

			if err := cli.usecase.CliConfigSetNewEUPFBaseURL(newBaseURL); err != nil {
				fmt.Fprintf(os.Stderr, "failed to set base URL: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Base URL updated successfully.")
		},
	}

	cmd.Flags().StringVar(&newBaseURL, "url", "", "New base URL to save in config (e.g. http://127.0.0.1:8081/api/v1)")
	return cmd
}

// config show-baseurl
func (cli *CLI) newConfigShowBaseURLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-baseurl",
		Short: "Show current base URL from config",
		Long: `Display the currently configured default eupf API base URL,
which is stored in the config-cli.yaml file under "eupf_addr".`,
		Run: func(cmd *cobra.Command, args []string) {
			url, err := cli.usecase.CliConfigShowEUPFBaseURL()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to retrieve base URL: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Current base URL: %s\n", url)
		},
	}
}
