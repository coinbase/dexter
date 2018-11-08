package tasks

import (
	"github.com/coinbase/dexter/util"

	log "github.com/Sirupsen/logrus"
)

//
// The init function should be used to declare your new task when Dexter starts up.
// Just pass a Task struct to the `add` function.
//
func init() {
	add(Task{
		// Give your task a short and descriptive name
		Name: "example-task",

		// Your task should also have a more detailed description that isn't too long
		Description: "this is a minimal task definition that can be used as an example",

		// If your task only makes sense in the context of some arguments, you can specify
		// a minimum number of arguments here.  Omit this if 0 arguments are ok.
		MinimumArguments: 0,

		// Define how many investigators need to sign an investigation containing this task
		ConsensusRequirement: 1,

		// supportedPlatforms contains valid values for go's runtime.GOOS
		// If this is omitted, the default value is all platforms.
		//
		// The list of valid platforms is here: https://github.com/golang/go/blob/master/src/go/build/syslist.go
		//
		supportedPlatforms: util.AllPlatforms,

		// Your action function is the actual code for the task.  It is defined below.
		actionFunction: exampleActionFunction,
	})
}

//
// This is the code that will execute when your task is run.
//
// If the platform Dexter is running on is not included in your
// supported platforms definition, this function will not be
// called, and an error will be logged.
//
// The arguments are an arbitrary-length slice of strings
// entered by the investigator who created the investigation.
// This lets you scope your task to something more specific,
// if desired.
//
// The ArtifactWriter will help you write data into the report
// that dexter will generate.  The first argument is the path
// within the report the artifact should be written to, and the
// second argument is the data.
//
func exampleActionFunction(arguments []string, writer *ArtifactWriter) {
	//
	// Logging is a good idea, and the logrus package makes
	// detailed logging easy!
	//
	log.WithFields(log.Fields{
		"at":        "tasks.exampleActionFunction",
		"path":      writer.path,
		"arguments": arguments,
	}).Info("running the example task")

	//
	// ...
	// This is all up to you!
	// ...
	//

	//
	// We can add a file to this report by using the "writeArtifact" function.
	//
	// Make sure that the file you create includes the prefix passed into this
	// function.  Let's create a simple hello world file as an example.
	//
	writer.Write(
		"hello_world.txt",
		[]byte("Hello, world!"),
	)

	// The `writeArtifact` function can also be used with a path, no need to
	// create the directories first.
	writer.Write(
		"artifacts/foo/bar/hello_world.txt",
		[]byte("Hello, world... again!"),
	)

	// Run into an error?  Include it in the report's errors.txt file
	writer.Error("something went wrong in this task")
}
