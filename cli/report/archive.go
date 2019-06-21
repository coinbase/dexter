package report

import (
	"os"
	"path"

	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func archiveReports(cmd *cobra.Command, args []string) {
	files, err := helpers.ListS3Path("reports/")
	if err != nil {
		color.HiRed("unable to list reports: " + err.Error())
		os.Exit(1)
	}

	for _, file := range files {
		base := path.Base(file)
		err = helpers.MoveS3File(file, "reports/_"+base)
		if err != nil {
			color.HiRed("error moving file for archive: " + err.Error())
			os.Exit(1)
		}
	}
}
