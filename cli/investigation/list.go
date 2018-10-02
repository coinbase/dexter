package investigation

import (
	"fmt"
	"os"
	"strings"

	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func listInvestigations(cmd *cobra.Command, args []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Investigation",
		"Issuer",
		"Tasks",
		"Scope",
		"Consensus",
		"Reviewed By",
	})

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
	)

	for _, inv := range engine.CurrentInvestigations() {
		table.Append([]string{
			inv.ID,
			inv.Issuer.Name,
			strings.Join(helpers.TaskStrings(inv.TaskList), ",\n"),
			strings.Join(inv.ScopeFactsStrings(), ",\n"),
			fmt.Sprintf("%d/%d", inv.ValidUniqueApprovers(), inv.MinimumConsensus()),
			strings.Join(inv.ApproverNames(), ",\n"),
		})
	}
	table.Render()
}
