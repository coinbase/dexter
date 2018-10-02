package report

import (
	"os"

	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func pruneReports(cmd *cobra.Command, args []string) {
	files, err := helpers.ListS3Path("reports/")
	if err != nil {
		color.HiRed("unable to list reports: " + err.Error())
		os.Exit(1)
	}

	for _, file := range files {
		err := helpers.DeleteS3File(file)
		if err != nil {
			color.HiRed(err.Error())
		}
	}
}
