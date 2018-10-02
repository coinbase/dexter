package investigator

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func createInvestigator(cmd *cobra.Command, args []string) {
	name := args[0] // Cobra ensures arg length, this is safe
	color.HiCyan("Initializing new investigator \"%s\" on local system...", name)
	privateKeyPEM, publicKey := generateInvestigatorKeys()
	writePrivateKey(name, privateKeyPEM)
	writeInvestigatorFile(name, publicKey)
}

func generateInvestigatorKeys() ([]byte, *rsa.PublicKey) {
	// Generate new RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		color.HiRed("fatal error generating RSA keys: %s", err.Error())
		os.Exit(1)
	}

	// Encode private key as password-protected PEM file
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err != nil {
		color.HiRed("fatal error encoding RSA private key: %s", err.Error())
		os.Exit(1)
	}
	pemBlock, err := x509.EncryptPEMBlock(rand.Reader, "ENCRYPTED PRIVATE KEY", privateBytes, []byte(cliutil.CollectNewPassword()), x509.PEMCipherAES128)
	if err != nil {
		color.HiRed("fatal error PEM encoding RSA private key: %s", err.Error())
		os.Exit(1)
	}
	blockBuffer := bytes.NewBuffer([]byte{})
	err = pem.Encode(blockBuffer, pemBlock)
	if err != nil {
		color.HiRed("fatal error PEM encoding RSA private key: %s", err.Error())
		os.Exit(1)
	}
	privateKeyFileData := blockBuffer.Bytes()

	// Generate the public key
	publicKey := privateKey.Public().(*rsa.PublicKey)

	return privateKeyFileData, publicKey
}

func writePrivateKey(name string, data []byte) {
	dexterDir := helpers.GetDexterDirectory()
	err := os.MkdirAll(filepath.FromSlash(dexterDir), 0700)
	if err != nil {
		color.HiRed("\nunable to create dexter config directory %s: %s", dexterDir, err.Error())
		os.Exit(1)
	}

	dexterKeyName := helpers.GetDexterKeyFile()
	if _, err := os.Stat(filepath.FromSlash(dexterKeyName)); err == nil {
		color.HiRed("\ndexter key file %s already exists", dexterKeyName)
		color.HiRed("If you would like to replace your key, please remove this file.")
		os.Exit(1)
	}
	ioutil.WriteFile(dexterKeyName, data, 0644)
}

func writeInvestigatorFile(name string, publicKey *rsa.PublicKey) {
	investigator := engine.Investigator{
		Name: name,
		PublicKey: engine.PublicKey{
			N: publicKey.N.String(),
			E: strconv.Itoa(publicKey.E),
		},
	}
	data, err := json.MarshalIndent(investigator, "", "  ")
	if err != nil {
		color.HiRed("fatal error encoding investigator definition: %s", err.Error())
		os.Exit(1)
	}

	// Save the local investigator file
	err = ioutil.WriteFile(helpers.GetDexterInvestigatorFile(), data, 0644)
	if err != nil {
		color.HiRed("fatal error writing local investigator definition: %s", err.Error())
		os.Exit(1)
	}

	// If investigators directory exists, place the public key there, as they are probably in the dexter
	// project directory and this makes creating the PR very easy.
	if stat, err := os.Stat("investigators"); err == nil && stat.Mode().IsDir() {
		err = ioutil.WriteFile("investigators/"+name+".json", data, 0644)
		if err != nil {
			color.HiRed("fatal error writing new investigator: %s", err.Error())
			os.Exit(1)
		}
		color.Yellow("%s.json has been generated in the investigators directory,", name)
		color.Yellow("submit this change as a pull request.")
	} else {
		err = ioutil.WriteFile(name+".json", data, 0644)
		if err != nil {
			color.HiRed("fatal error writing new investigator: %s", err.Error())
			os.Exit(1)
		}
		color.Yellow("%s.json has been generated in the current directory.", name)
		color.Yellow("This file should be moved into dexter's investigators")
		color.Yellow("directory in a pull request.")
	}
}
