package investigation

import (
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "investigation [cmd]",
	Short: "Manage investigations",
	Long:  `This command is used to manage investigations.`,
	Args:  cobra.MinimumNArgs(1),
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new dexter investigation",
	Long:  `Create a new dexter investigation and upload it for approval`,
	Args:  cobra.MaximumNArgs(0),
	Run:   createInvestigation,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Dexter investigations",
	Long:  `Print a list of all dexter investigations`,
	Args:  cobra.MaximumNArgs(0),
	Run:   listInvestigations,
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Permanently delete all investigations",
	Long:  `Download an archive of investigations and empty the investigations bucket`,
	Args:  cobra.MaximumNArgs(0),
	Run:   pruneInvestigations,
}

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Sign pending investigations for consensus",
	Long:  `Print the details of an investigation and provide a signature for consensus`,
	Args:  cobra.MinimumNArgs(1),
	Run:   approveInvestigation,
}

func CommandSuite() *cobra.Command {
	cmd.AddCommand(createCmd)
	cmd.AddCommand(listCmd)
	cmd.AddCommand(pruneCmd)
	cmd.AddCommand(approveCmd)
	return cmd
}
