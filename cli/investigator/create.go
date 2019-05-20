package investigator

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func createInvestigator(cmd *cobra.Command, args []string) {
	name := args[0] // Cobra ensures arg length, this is safe
	color.HiCyan("Initializing new investigator \"%s\" on local system...", name)
	investigator, privateKeyPEM, err := engine.NewInvestigator(name, cliutil.CollectNewPassword())
	if err != nil {
		color.HiRed(err.Error())
		os.Exit(1)
	}
	writePrivateKey(investigator, privateKeyPEM)
	writePublicKey(investigator)
}

func writePrivateKey(investigator engine.Investigator, privatePEM []byte) {
	// Create dexter directory if needed
	dexterDir := helpers.GetDexterDirectory()
	err := os.MkdirAll(filepath.FromSlash(dexterDir), 0700)
	if err != nil {
		color.HiRed("\nunable to create dexter config directory %s: %s", dexterDir, err.Error())
		os.Exit(1)
	}

	// Serialize the investigator
	investigatorData, err := investigator.String()
	if err != nil {
		color.HiRed("fatal error serializing new investigator: %s", err.Error())
		os.Exit(1)
	}

	// Save the local investigator file
	err = ioutil.WriteFile(helpers.GetDexterInvestigatorFile(), investigatorData, 0644)
	if err != nil {
		color.HiRed("fatal error writing local investigator definition: %s", err.Error())
		os.Exit(1)
	}

	// Save the local private key
	dexterKeyName := helpers.GetDexterKeyFile()
	if _, err = os.Stat(filepath.FromSlash(dexterKeyName)); err == nil {
		color.HiRed("\ndexter key file %s already exists", dexterKeyName)
		color.HiRed("If you would like to replace your key, please remove this file.")
		os.Exit(1)
	}
	err = ioutil.WriteFile(dexterKeyName, privatePEM, 0644)
	if err != nil {
		color.HiRed("fatal error writing investigator private key: " + err.Error())
		os.Exit(1)
	}
}

func writePublicKey(investigator engine.Investigator) {
	data, err := json.MarshalIndent(investigator, "", "  ")
	if err != nil {
		color.HiRed("fatal error encoding investigator: " + err.Error())
		os.Exit(1)
	}

	err = ioutil.WriteFile(investigator.Name+".json", data, 0644)
	if err != nil {
		color.HiRed("fatal error writing investigator file: " + err.Error())
		os.Exit(1)
	}
	color.Green("New investigator file created: " + investigator.Name + ".json")
	color.Yellow("This must be uploaded to Dexter by your Dexter administrator.")
}
