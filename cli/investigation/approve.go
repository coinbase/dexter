package investigation

import (
	"os"
	"strconv"
	"strings"

	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func approveInvestigation(cmd *cobra.Command, args []string) {
	uuid := args[0]
	inv, err := engine.InvestigationByID(uuid)
	if err != nil {
		color.HiRed("error looking up investigation: " + err.Error())
		return
	}

	color.HiYellow("Provide your password to approve the following investigation:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.FgHiMagentaColor},
	)
	table.Append([]string{"ID", inv.ID})
	table.Append([]string{"Issued By", inv.Issuer.Name})
	table.Append([]string{"Tasks", strings.Join(helpers.TaskStrings(inv.TaskList), ", ")})
	table.Append([]string{"Scope", strings.Join(inv.ScopeFactsStrings(), ", ")})
	table.Append([]string{"Kill Containers?", strconv.FormatBool(inv.KillContainers)})
	table.Append([]string{"Kill Host?", strconv.FormatBool(inv.KillHost)})
	table.Append([]string{"Recipients", strings.Join(inv.RecipientNames, ", ")})
	table.Append([]string{"Approvers", strings.Join(inv.ApproverNames(), ", ")})
	table.Render()

	inv.Approve(helpers.LoadLocalKey(cliutil.CollectPassword))
	err = inv.Upload()
	if err != nil {
		color.HiRed("Failed to upload approval: " + err.Error())
		return
	}
	success := color.New(color.FgHiGreen, color.Bold)
	success.Println("Investigation Approved!")
}
