//
// The daemon package processes the command line entrypoint for
// the dexter daemon.
//
package daemon

import (
	"github.com/coinbase/dexter/engine"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "daemon",
	Short: "Launch Dexter Daemon",
	Long: `This command is used to start Dexter on hosts that will be
receiving investigations to run.
`,
	Args: cobra.MaximumNArgs(0),
	Run:  startEngine,
}

//
// Return the set of cobra commands used for the daemon subcommand
//
func CommandSuite() *cobra.Command {
	return cmd
}

func startEngine(_ *cobra.Command, _ []string) {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg: "message",
		},
	})
	log.WithFields(log.Fields{
		"at": "daemon.startEngine",
	}).Info("Starting Dexter Daemon")

	engine.Start()
}
