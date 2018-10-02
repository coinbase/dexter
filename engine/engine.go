//
// The Dexter engine contains all the functionality to run the dexter daemon loop.
//
package engine

import (
	log "github.com/Sirupsen/logrus"
)

//
// Poll for investigations, validate them, and run the tasks if in scope.
//
func Start() {
	for investigation := range NewS3Poller().Poll() {
		err := investigation.validate()
		if err != nil {
			log.WithFields(log.Fields{
				"at":            "engine.Start",
				"investigation": investigation.ID,
			}).Error(err)
			continue
		}

		investigation.run()
		investigation.report()
		investigation.cleanup()
	}
}
