package investigator

import (
	"strings"

	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func revokeInvestigator(cmd *cobra.Command, args []string) {
	for _, name := range args {
		color.HiCyan("Revoking investigator \"%s\" ", name)
		path := "investigators/" + name + ".json"
		err := helpers.DeleteS3File(path)
		if err != nil {
			color.HiRed("error revoking investigator: " + err.Error())
		} else {
			color.HiGreen("Investigator Revoked!")
		}

		color.Yellow("Deleting all old reports for %s", name)
		files, err := helpers.ListS3Path("reports/")
		if err != nil {
			color.HiRed("error listing reports: %s", err)
		}
		for _, file := range files {
			if strings.Contains(file, name) {
				err := helpers.DeleteS3File(file)
				if err != nil {
					color.HiRed(err.Error())
				}
			}
		}
	}
}
