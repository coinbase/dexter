package investigation

import (
	"github.com/spf13/cobra"
)

var showArchived bool

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

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Hide investigations from Dexter",
	Long:  `Mark the investigations as archived, removing them from the Dexter cli while preserving them on S3`,
	Args:  cobra.MaximumNArgs(0),
	Run:   archiveInvestigations,
}

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Sign pending investigations for consensus",
	Long:  `Print the details of an investigation and provide a signature for consensus`,
	Args:  cobra.MinimumNArgs(1),
	Run:   approveInvestigation,
}

func CommandSuite() *cobra.Command {
	listCmd.PersistentFlags().BoolVar(&showArchived, "archived", false, "show archived investigations")

	cmd.AddCommand(createCmd)
	cmd.AddCommand(listCmd)
	cmd.AddCommand(archiveCmd)
	cmd.AddCommand(approveCmd)
	return cmd
}
