//
// The investigator package contains command line tools for creating
// and revoking investigators.
//
package investigator

import (
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "investigator [cmd]",
	Short: "Manage investigators",
	Long:  `This command is used to manage investigators.`,
	Args:  cobra.MinimumNArgs(1),
}

var createCmd = &cobra.Command{
	Use:   "init [username]",
	Short: "Create a new dexter investigator",
	Long: `This command creates a new investigator for the local machine.

A RSA key pair is generated which is used to sign investigations.
This is saved to the local filesystem, and a file is generated in
the current working directory which can be submitted in a pull
request to Dexter to add this investigator.`,
	Args: cobra.MinimumNArgs(1),
	Run:  createInvestigator,
}

var revokeCmd = &cobra.Command{
	Use:   "revoke [username] <usernames...>",
	Short: "Revoke dexter investigators",
	Long:  "Remove investigator(s) from Dexter and delete all reports for the investigator(s).  Reports will be preseved for other users.",
	Args:  cobra.MinimumNArgs(1),
	Run:   revokeInvestigator,
}

func CommandSuite() *cobra.Command {
	cmd.AddCommand(createCmd)
	cmd.AddCommand(revokeCmd)
	return cmd
}
