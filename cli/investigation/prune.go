package investigation

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Download investigation files from the S3 for archiving, then remove them
// from the directory.
func pruneInvestigations(cmd *cobra.Command, args []string) {
	filenames, err := helpers.ListS3Path("investigations/")
	if err != nil {
		color.HiRed("unable to list investigations: " + err.Error())
		os.Exit(1)
	}

	// Create archive directory
	err = os.MkdirAll("InvestigationArchive", 0700)
	if err != nil {
		color.HiRed("unable to create directory for investigation archive: " + err.Error())
		os.Exit(1)
	}

	for _, filename := range filenames {
		// Download into an archive directory
		data, err := helpers.GetS3File(filename)
		if err != nil {
			color.HiRed("unable to download archive " + filename + ": " + err.Error())
			continue
		}
		archive := "InvestigationArchive/" + path.Base(filename)
		ferr := ioutil.WriteFile(archive, data, 0644)
		if ferr != nil {
			color.HiRed("unable to write archive " + archive + ": " + ferr.Error())
			continue
		}

		// Delete the object from S3
		derr := helpers.DeleteS3File(filename)
		if derr != nil {
			color.HiRed("unable to delete " + filename + ": " + derr.Error())
		}
	}
}
