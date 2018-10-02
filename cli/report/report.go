//
// The report package contains the command line tools for interacting with reports.
//
package report

import (
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "report [cmd]",
	Short: "Manage reports",
	Long:  `This command is used to manage reports.`,
	Args:  cobra.MinimumNArgs(1),
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Dexter reports",
	Long:  `Print a list of dexter reports that are available for download.`,
	Args:  cobra.MaximumNArgs(0),
	Run:   listReports,
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete all report files",
	Long:  `Delete all report files`,
	Args:  cobra.MaximumNArgs(0),
	Run:   pruneReports,
}

var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Download and decrypt a report",
	Long:  `Download a report and decrypt it into a local directory`,
	Args:  cobra.MinimumNArgs(1),
	Run:   retrieveReport,
}

func CommandSuite() *cobra.Command {
	cmd.AddCommand(listCmd)
	cmd.AddCommand(pruneCmd)
	cmd.AddCommand(retrieveCmd)
	return cmd
}
