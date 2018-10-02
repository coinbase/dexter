package engine

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/coinbase/dexter/embedded"
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/fatih/color"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
)

//
// Values for RSA public key, represented as strings for JSON.
//
type PublicKey struct {
	N, E string
}

//
// An investigator is defined by their name and public key.
//
type Investigator struct {
	PublicKey PublicKey
	Name      string
}

//
// Lookup an embedded investigator and parse their public key into
// an *rsa.PublicKey.
//
func GetPublicKeyForInvestigator(name string) (*rsa.PublicKey, error) {
	set := LoadInvestigators()
	for _, investigator := range set {
		if investigator.Name == name {
			n := new(big.Int)
			n, ok := n.SetString(investigator.PublicKey.N, 10)
			if !ok {
				log.WithFields(log.Fields{
					"investigator": investigator.Name,
					"N":            investigator.PublicKey.N,
					"at":           "engine.getPublicKeyForInvestigator",
				}).Error("error parsing N value")
				return &rsa.PublicKey{}, errors.New("error parsing N value")
			}
			e, err := strconv.Atoi(investigator.PublicKey.E)
			if err != nil {
				log.WithFields(log.Fields{
					"at":    "engine.getPublicKeyForInvestigator",
					"error": err.Error(),
				}).Error("error parsing E value")
				return &rsa.PublicKey{}, errors.New("error parsing E value")
			}
			return &rsa.PublicKey{
				N: n,
				E: e,
			}, nil
		}
	}
	log.WithFields(log.Fields{
		"at":   "engine.getPublicKeyForInvestigator",
		"name": name,
	}).Fatal("named investigator not found")
	return &rsa.PublicKey{}, errors.New("named investigator not found")
}

//
// Return the list of embedded investigators.
//
func LoadInvestigatorNames() (list []string) {
	set := LoadInvestigators()
	for _, member := range set {
		list = append(list, member.Name)
	}
	return
}

//
// Return the local investigator as an Investigator struct.
//
func LoadLocalInvestigator() Investigator {
	data, err := ioutil.ReadFile(helpers.GetDexterInvestigatorFile())
	if err != nil {
		color.HiRed("error reading investigator file: " + err.Error())
		os.Exit(1)
	}
	var investigator Investigator
	err = json.Unmarshal(data, &investigator)
	if err != nil {
		color.HiRed("error parsing investigator file: " + err.Error())
		os.Exit(1)
	}
	return investigator
}

//
// Return the name of the investigator currently operating
// Dexter from the CLI.
//
func LocalInvestigatorName() string {
	person := LoadLocalInvestigator()
	return person.Name
}

//
// Load the investigator structs from the embedded files and
// return a slice of investigators.
//
func LoadInvestigators() (list []Investigator) {
	investigatorList, err := embedded.WalkDirs("investigators", true)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "engine.LoadInvestigators",
			"error": err.Error(),
		}).Fatal("unable to walk embedded investigator list")
	}
	for _, filename := range investigatorList {
		investigatorJSON, err := embedded.ReadFile(filename)
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.LoadInvestigators",
				"error": err.Error(),
				"name":  filename,
			}).Error("unable to read investigator file data")
			continue
		}
		person := Investigator{}
		err = json.Unmarshal(investigatorJSON, &person)
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "engine.LoadInvestigators",
				"error": err.Error(),
				"name":  filename,
			}).Error("error parsing investigator struct")
		} else {
			list = append(list, person)
		}
	}

	if len(list) == 0 {
		log.WithFields(log.Fields{
			"at": "engine.LoadInvestigators",
		}).Fatal("no investigators loaded")
	}
	return
}
