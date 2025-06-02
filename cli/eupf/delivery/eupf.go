package delivery

import (
	"fmt"
	"os"

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

	root.AddCommand(cli.newConfigCmd())
	root.AddCommand(cli.newSessionCmd())
	root.AddCommand(cli.newCliConfigCmd())
	root.AddCommand(cli.newTraceCmd())
	root.AddCommand(cli.newBackupCmd())
	root.AddCommand(cli.newCompletion())

	return cli
}

func (cli *CLI) Execute() error {
	return cli.Root.Execute()
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
