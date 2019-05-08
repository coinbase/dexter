package tasks

import (
	"io/ioutil"

	"github.com/coinbase/dexter/util"

	log "github.com/Sirupsen/logrus"
)

func init() {
	add(Task{
		Name:                 "get-file",
		Description:          "retrieve files from host",
		MinimumArguments:     1,
		ConsensusRequirement: 1,
		supportedPlatforms:   util.AllPlatforms,
		actionFunction:       getFile,
	})
}

func getFile(arguments []string, writer *ArtifactWriter) {

	log.WithFields(log.Fields{
		"at":        "tasks.getFile",
		"path":      writer.path,
		"arguments": arguments,
	}).Info("retrieving files")

	for _, arg := range arguments {
		bytes, err := ioutil.ReadFile(arg)
		if err != nil {
			log.WithFields(log.Fields{
				"at":   "tasks.getFile",
				"path": writer.path,
				"file": arg,
			}).Error("error reading file")
			writer.Error("error reading file: " + arg)
		} else {
			writer.Write(arg, bytes)
		}
	}
}
