package tasks

import (
	"github.com/coinbase/dexter/embedded"

	log "github.com/Sirupsen/logrus"
)

func init() {
	add(Task{
		Name:             "revoke-key",
		Description:      "invalidate investigator key on all instances of dexter",
		actionFunction:   revokeInvestigatorKey,
		MinimumArguments: 1,
	})
}

//
// Overwrite the public key file for the named investigators,
// making it impossible for Dexter to generate a report that
// the investigators can read.  This is not permanent, the
// keys must also be removed from the Dexter repo, otherwise
// they will return next deploy, but this will revoke all the
// keys in the Dexters that are already out there.
//
func revokeInvestigatorKey(args []string, writer *ArtifactWriter) {
	if len(args) < 1 {
		errstr := "no user specified to revoke"
		log.WithFields(log.Fields{
			"at": "tasks.revokeInvestigatorKey",
		}).Error(errstr)
		writer.Error(errstr)
		return
	}
	for _, user := range args {
		err := embedded.WriteFile("investigators/"+user+".json", []byte{}, 0777)
		if err == nil {
			log.WithFields(log.Fields{
				"at":           "tasks.revokeInvestigatorKey",
				"investigator": user,
			}).Info("destroyed investigator key")
		} else {
			errstr := "error destroying investigator key"
			log.WithFields(log.Fields{
				"at":           "tasks.revokeInvestigatorKey",
				"investigator": user,
				"error":        err.Error(),
			}).Info(errstr)
			writer.Error(errstr + " for " + user + ": " + err.Error())
		}
	}
}
