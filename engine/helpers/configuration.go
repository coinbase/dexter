package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/coinbase/dexter/embedded"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
)

var osquerySocket string
var pollInterval int
var stubbedProjectName string

//
// The structure for the Dexter JSON config file
//
type DexterConfig struct {
	ConsensusRequirements map[string]int
	ProjectName           ProjectNameConfig
	OSQuerySocket         string
	S3Bucket              string
	PollIntervalSeconds   int
}

//
// A description of how Dexter can find the project name of a host.
//
// Type can define "file" or "envar", and Location can refer to
// the path or envar name.
//
type ProjectNameConfig struct {
	Type     string
	Location string
}

var s3String *string

//
// Return an AWS string containing the S3 bucket defined by the DEXTER_AWS_S3_BUCKET environment variable
//
func S3Bucket() *string {
	if LocalDemoPath != "" {
		return aws.String("local://" + LocalDemoPath)
	}
	if s3String == nil {
		s3String = aws.String(os.Getenv("DEXTER_AWS_S3_BUCKET"))
		if *s3String == "" {
			configBucket := LoadDexterConfig().S3Bucket
			s3String = &configBucket
			if *s3String == "" {
				log.WithFields(log.Fields{
					"at": "helpers.S3Bucket",
				}).Fatal("dexter bucket not specified")
			}
		}
	}
	return s3String
}

//
// Return the section of the config that describes task consensus requirements.
//
func ConsensusSet() map[string]int {
	return LoadDexterConfig().ConsensusRequirements
}

//
// Lookup and cache the pool interval
//
func PollInterval() int {
	if pollInterval > 0 {
		return pollInterval
	}
	pollInterval = LoadDexterConfig().PollIntervalSeconds
	return pollInterval
}

//
// Lookup and cache the osquery socket
//
func OSQuerySocket() string {
	if osquerySocket != "" {
		return osquerySocket
	}
	osquerySocket = LoadDexterConfig().OSQuerySocket
	return osquerySocket
}

//
// Load the embedded config file and return the parsed structure.
//
func LoadDexterConfig() DexterConfig {
	dexterConfigJSON, err := embedded.ReadFile("config/dexter.json")
	if err != nil {
		log.WithFields(log.Fields{
			"at": "helpers.LoadDexterConfig",
		}).Fatal("embedded dexter config not present")
	}
	dexterConfig := DexterConfig{}
	err = json.Unmarshal(dexterConfigJSON, &dexterConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.LoadDexterConfig",
			"error": err.Error(),
		}).Fatal("embedded dexter config contains invalid JSON")
	}
	return dexterConfig
}

//
// Stub all calls to ProjectName with a string, for testing.
//
func StubProjectName(str string) {
	stubbedProjectName = str
}

//
// Get the project name for this host.  Useful when scoping in a production environment.
//
func ProjectName() string {
	if stubbedProjectName != "" {
		return stubbedProjectName
	}
	config := LoadDexterConfig().ProjectName
	if config.Type == "file" {
		data, err := ioutil.ReadFile(config.Location)
		if err != nil {
			log.WithFields(log.Fields{
				"at":       "helpers.ProjectName",
				"filename": config.Location,
				"error":    err.Error(),
			}).Fatal("unable to read project name file")
		}
		return string(data)
	} else if config.Type == "envar" {
		return os.Getenv(config.Location)
	}
	log.Fatal("no source for project name configured")
	return ""
}
