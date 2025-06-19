package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
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
	var formatJson bool
	var verbose bool

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

			if formatJson {
				jsonBytes, err := json.MarshalIndent(sessions, "", "    ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to marshal PFCP sessions: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(string(jsonBytes))
				return
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Local SEID", "Remote SEID", "MSISDN", "IMSI"})
			for _, s := range sessions {
				t.AppendRow([]any{s.LocalSEID, s.RemoteSEID, s.MSISDN, s.IMSI})
			}

			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 1, Transformer: formatAsHex},
				{Number: 2, Transformer: formatAsHex},
			})

			t.SetStyle(table.StyleLight)
			t.Render()

			if len(sessions) == 1 || verbose {
				t = table.NewWriter()
				t.SetOutputMirror(os.Stdout)

				t.AppendHeader(table.Row{"SEID", "PDR ID", "TEID", "IPv4", "IPv6", "FAR ID", "QER ID", "URR1 ID", "URR2 ID"})
				for _, s := range sessions {
					for _, pdr := range s.PDRs {
						t.AppendRow([]any{
							s.LocalSEID,
							pdr.PdrID,
							pdr.Teid,
							pdr.Ipv4,
							pdr.Ipv6,
							pdr.PdrInfo.FarId,
							pdr.PdrInfo.QerId,
							pdr.PdrInfo.Urr1Id,
							pdr.PdrInfo.Urr2Id,
						})
					}
				}

				t.SetColumnConfigs([]table.ColumnConfig{
					{Number: 1, AutoMerge: true, Transformer: formatAsHex},
				})

				t.SetStyle(table.StyleLight)
				t.Render()

			}

		},
	}

	cmd.Flags().StringVar(&ip, "ip", "", "Filter by IP address")
	cmd.Flags().StringVarP(&teid, "teid", "t", "", "Filter by TEID")
	cmd.Flags().BoolVarP(&formatJson, "json", "j", false, "Output the result in JSON format")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed session information")
	cmd.Flags().StringVar(&baseURL, "baseurl", "", "Optional base URL to override eupf API address (e.g., http://localhost:8081/api/v1)")

	return cmd
}

func formatAsHex(val any) string {
	switch v := val.(type) {
	case uint64:
		return fmt.Sprintf("%#x", v)
	default:
		return fmt.Sprintf("%v", v)
	}
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
