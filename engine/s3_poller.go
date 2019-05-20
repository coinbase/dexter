package engine

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/util"
)

//
// A poller that will stream new files from the Dexter
// investigations S3 bucket.
//
type S3Poller struct{}

var seenFiles = make([]string, 0)

//
// Create a new S3 poller.
//
func NewS3Poller() *S3Poller {
	return &S3Poller{}
}

//
// Get a chan of investigation structs from the Dexter investigations S3 bucket.
//
func (poller *S3Poller) Poll() chan Investigation {
	newInvestigations := make(chan Investigation)
	go pollInvestigations(newInvestigations)
	return newInvestigations
}

func pollInvestigations(newInvestigations chan Investigation) {
	investigations, err := helpers.ListS3Path("investigations/")
	if err != nil {
		log.WithFields(log.Fields{
			"at":     "engine.pollInvestigations",
			"error":  err.Error(),
			"bucket": helpers.S3Bucket(),
		}).Fatal("error listing investigation objects in bucket")
	}

	seenFiles = append(seenFiles, investigations...)

	for {
		investigations, err = helpers.ListS3Path("investigations/")
		if err != nil {
			log.WithFields(log.Fields{
				"at":     "engine.pollInvestigations",
				"error":  err.Error(),
				"bucket": helpers.S3Bucket(),
			}).Error("error listing investigation objects in bucket")
			time.Sleep(10 * time.Second)
			continue
		}
		for _, key := range changes(investigations) {
			data, err := helpers.GetS3File(key)
			if err != nil {
				log.WithFields(log.Fields{
					"at":     "engine.pollInvestigations",
					"error":  err.Error(),
					"bucket": helpers.S3Bucket(),
					"key":    key,
				}).Error("error getting investigation object from bucket")
				continue
			}
			var inv = Investigation{}
			err = json.Unmarshal(data, &inv)
			if err != nil {
				log.WithFields(log.Fields{
					"at":    "engine.pollInvestigations",
					"error": err.Error(),
					"key":   key,
				}).Error("downloaded json-invalid investigation")
			} else {
				newInvestigations <- inv
			}
		}
		time.Sleep(time.Duration(helpers.PollInterval()) * time.Second)
	}
}

func changes(set []string) []string {
	unseen := make([]string, 0)
	for _, member := range set {
		if !util.StringsInclude(seenFiles, member) {
			unseen = append(unseen, member)
			seenFiles = append(seenFiles, member)
		}
	}
	return unseen
}
