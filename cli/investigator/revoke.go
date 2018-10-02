package investigator

import (
	"strings"

	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func revokeInvestigator(cmd *cobra.Command, args []string) {
	name := args[0] // Cobra ensures arg length, this is safe
	color.HiCyan("Revoking investigator \"%s\" ", name)

	investigation := engine.Investigation{
		ID: helpers.NewDexterID(),
		TaskList: map[string][]string{
			"revoke-key": []string{name},
		},
		KillContainers: false,
		KillHost:       false,
		RecipientNames: []string{},
		Issuer:         engine.Signature{Name: engine.LocalInvestigatorName()},
	}
	color.Yellow("Please provide your password to sign this request:")
	investigation.Sign(helpers.LoadLocalKey(cliutil.CollectPassword))
	err := investigation.Upload()
	if err != nil {
		color.HiRed("error uploading investigation: " + err.Error())
	} else {
		color.HiGreen("Investigation Uploaded!")
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
