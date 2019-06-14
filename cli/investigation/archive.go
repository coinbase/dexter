package investigation

import (
	"os"
	"path"

	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Add an underscore to investigation files to hide them by default
func archiveInvestigations(cmd *cobra.Command, args []string) {
	filenames, err := helpers.ListS3Path("investigations/")
	if err != nil {
		color.HiRed("unable to list investigations: " + err.Error())
		os.Exit(1)
	}

	for _, filename := range filenames {
		base := path.Base(filename)
		err = helpers.MoveS3File(filename, "investigations/_"+base)
		if err != nil {
			color.HiRed("error moving file for archive: " + err.Error())
			os.Exit(1)
		}
	}
}
