//
// The report package contains the command line tools for interacting with reports.
//
package report

import (
	"github.com/spf13/cobra"
)

var showArchived bool

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

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive all reports",
	Long:  `Mark reports as archived, removing them from the Dexter cli while preserving them in S3`,
	Args:  cobra.MaximumNArgs(0),
	Run:   archiveReports,
}

var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Download and decrypt a report",
	Long:  `Download a report and decrypt it into a local directory`,
	Args:  cobra.MinimumNArgs(1),
	Run:   retrieveReport,
}

func CommandSuite() *cobra.Command {
	listCmd.PersistentFlags().BoolVar(&showArchived, "archived", false, "show archived reports")

	cmd.AddCommand(listCmd)
	cmd.AddCommand(archiveCmd)
	cmd.AddCommand(retrieveCmd)
	return cmd
}
