package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func (cli *CLI) newSessionCmd() *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Pfcp session management",
	}

	sessionCmd.AddCommand(cli.newSessionShowCmd())
	sessionCmd.AddCommand(cli.newSessionReleaseCmd())

	return sessionCmd
}

// session show -ip <ip> -t <teid>
func (cli *CLI) newSessionShowCmd() *cobra.Command {
	var baseURL string
	var ip string
	var teid string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show PFCP sessions filtered by IP or TEID",
		Long: `Show all PFCP sessions or filter by IP address or TEID.
If no parameters are provided, all sessions are listed.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			sessions, err := cli.usecase.SessionShow(ctx, ip, teid, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to retrieve PFCP sessions: %v\n", err)
				os.Exit(1)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found")
				return
			}

			jsonBytes, err := json.MarshalIndent(sessions, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal PFCP sessions: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("PFCP Sessions JSON:", string(jsonBytes))
		},
	}

	cmd.Flags().StringVar(&ip, "ip", "", "Filter by IP address")
	cmd.Flags().StringVarP(&teid, "teid", "t", "", "Filter by TEID")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

// session release -i <imsi> -m <msisdn> -s <session-id>
func (cli *CLI) newSessionReleaseCmd() *cobra.Command {
	var baseURL string
	var imsi string
	var msisdn string
	var sessionID string

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Delete PFCP session by IMSI, MSISDN, or Session ID",
		Long: `Delete a PFCP session. You must specify exactly one of --imsi, --msisdn, or --session-id.
Use --baseurl to override the eupf API endpoint.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			count := 0
			if imsi != "" {
				count++
			}
			if msisdn != "" {
				count++
			}
			if sessionID != "" {
				count++
			}

			if count != 1 {
				fmt.Fprintln(os.Stderr, "error: you must specify exactly one of --imsi, --msisdn, or --session-id")
				os.Exit(1)
			}

			err := cli.usecase.SessionRelease(ctx, imsi, msisdn, sessionID, baseURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to release PFCP session: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("PFCP session release command executed successfully.")
		},
	}

	cmd.Flags().StringVarP(&imsi, "imsi", "i", "", "IMSI of the session to release")
	cmd.Flags().StringVarP(&msisdn, "msisdn", "m", "", "MSISDN of the session to release")
	cmd.Flags().StringVarP(&sessionID, "session-id", "s", "", "Session ID to release")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}
