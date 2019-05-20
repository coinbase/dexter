package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/fatih/color"

	"github.com/coinbase/dexter/util"
)

var keyCached = false
var cachedKey *rsa.PrivateKey

//
// Find the configuration directory for Dexter.
//
func GetDexterDirectory() string {
	usr, err := user.Current()
	if err != nil {
		color.HiRed("\nunable to get current home directory: %s", err.Error())
		os.Exit(1)
	}
	return usr.HomeDir + "/.dexter"
}

//
// Return the full path for the file that stores a user's private key.
//
func GetDexterKeyFile() string {
	return GetDexterDirectory() + "/key.pem"
}

//
// Return the full path for the file that stores the local investigator data.
//
func GetDexterInvestigatorFile() string {
	return GetDexterDirectory() + "/investigator.json"
}

//
// Load the local investigator's private key and decrypt it by getting the password
// from user interaction.
//
func LoadLocalKey(passwordRetriever func() string) *rsa.PrivateKey {
	if keyCached {
		return cachedKey
	}

	keyPEM, err := ioutil.ReadFile(GetDexterKeyFile())
	if err != nil {
		color.HiRed("Error reading local key file: " + err.Error())
		os.Exit(1)
	}

	block, _ := pem.Decode(keyPEM)
	for {
		keyDer, err := x509.DecryptPEMBlock(block, []byte(passwordRetriever()))
		if err != nil {
			color.Red("Decryption error in x509.DecryptPEMBlock: " + err.Error())
			color.HiRed("Try again")
			continue
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(keyDer)
		if err != nil {
			color.Red("Parsing error in x509.ParsePKCS1PrivateKey: " + err.Error())
			os.Exit(1)
			continue
		}
		cachedKey = privateKey
		keyCached = true
		return privateKey
	}
}

//
// Given a prefix for an ID in Dexter, return the full ID if there is enough
// specificity.  If there is too much ambiguity in the ID, and there are
// multiple possible matches, return an error.  This function works for both
// investigation and report IDs.
//
func ResolveUUID(partial string) (string, error) {
	allUUIDs := investigationUUIDs()
	allUUIDs = util.AppendUnique(allUUIDs, reportUUIDs())

	possibleMatches := make([]string, 0)
	for _, uuid := range allUUIDs {
		if strings.HasPrefix(uuid, partial) {
			possibleMatches = append(possibleMatches, uuid)
		}
	}

	if len(possibleMatches) == 1 {
		return possibleMatches[0], nil
	} else if len(possibleMatches) > 1 {
		return "", errors.New("too many possible UUID matches")
	}
	return "", errors.New("no possible UUID matches")
}

func investigationUUIDs() []string {
	investigations, err := ListS3Path("investigations/")
	if err != nil {
		color.HiRed(err.Error())
	}

	seenUUIDs := make(map[string]bool)
	allUUIDs := make([]string, 0)
	for _, filename := range investigations {
		uuid := filename[15:23]
		if _, present := seenUUIDs[uuid]; !present {
			seenUUIDs[uuid] = true
			allUUIDs = append(allUUIDs, uuid)
		}
	}
	return allUUIDs
}

func reportUUIDs() []string {
	reports, err := ListS3Path("reports/")
	if err != nil {
		color.HiRed(err.Error())
	}

	seenUUIDs := make(map[string]bool)
	allUUIDs := make([]string, 0)
	for _, filename := range reports {
		uuid := filename[8:16]
		if _, present := seenUUIDs[uuid]; !present {
			seenUUIDs[uuid] = true
			allUUIDs = append(allUUIDs, uuid)
		}
	}
	return allUUIDs
}
