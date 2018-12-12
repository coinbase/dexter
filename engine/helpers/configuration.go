package helpers

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
)

var osquerySocket string
var pollInterval int
var stubbedProjectName string
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
			log.WithFields(log.Fields{
				"at": "helpers.S3Bucket",
			}).Fatal("dexter bucket not specified")
		}
	}
	return s3String
}

//
// Lookup and cache the pool interval
//
func PollInterval() int {
	if pollInterval > 0 {
		return pollInterval
	}

	envarName := "DEXTER_POLL_INTERVAL_SECONDS"
	intervalStr := os.Getenv(envarName)
	if intervalStr == "" {
		log.WithFields(log.Fields{
			"at":    "helpers.PollInterval",
			"envar": envarName,
		}).Warn("poll interval envar not set, using 10 seconds")
		pollInterval = 10
		return 10
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.PollInterval",
			"error": err.Error(),
			"value": intervalStr,
		}).Warn("unable to convert interval value to int, using 10 seconds")
		pollInterval = 10
		return 10
	}

	pollInterval = interval
	return pollInterval
}

//
// Lookup and cache the osquery socket
//
func OSQuerySocket() string {
	if osquerySocket != "" {
		return osquerySocket
	}

	defaultSocket := "/var/osquery/osquery.em"
	osqueryEnvar := "DEXTER_OSQUERY_SOCKET"
	osquerySocket = os.Getenv(osquerySocket)

	if osquerySocket == "" {
		log.WithFields(log.Fields{
			"at":      "helpers.OSQuerySocket",
			"default": "defaultSocket",
			"envar":   osqueryEnvar,
		}).Warn("no osquery socket defined in envar, using default")
		osquerySocket = defaultSocket
	}

	return osquerySocket
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
	projectEnvar := os.Getenv("DEXTER_PROJECT_NAME_CONFIG")
	if projectEnvar == "" {
		log.WithFields(log.Fields{
			"at": "helpers.ProjectName",
		}).Warn("no project name configured, project name facts will not work")
		return ""
	}
	if strings.HasPrefix(projectEnvar, "file://") {
		location := strings.TrimPrefix(projectEnvar, "file://")
		data, err := ioutil.ReadFile(location)
		if err != nil {
			log.WithFields(log.Fields{
				"at":       "helpers.ProjectName",
				"filename": location,
				"error":    err.Error(),
			}).Fatal("unable to read project name file")
		}
		return string(data)
	} else if strings.HasPrefix(projectEnvar, "envar://") {
		location := strings.TrimPrefix(projectEnvar, "envar://")
		return os.Getenv(location)
	}

	log.WithFields(log.Fields{
		"at":     "helpers.ProjectName",
		"config": projectEnvar,
	}).Fatal("error parsing project name configuration")
	return ""
}
