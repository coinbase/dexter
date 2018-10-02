//
// The tasks package contains all the tasks Dexter can run.
//
// To add a new task, start by making a copy of the example file.
//
package tasks

import (
	"github.com/coinbase/dexter/util"

	log "github.com/Sirupsen/logrus"

	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

//
// A Task defines something Dexter can do on a host.
//
type Task struct {
	Name               string
	Description        string
	MinimumArguments   int
	supportedPlatforms []string
	actionFunction
}

// An ArtifactWriter helps you create files in the correct
// path for a report.
type ArtifactWriter struct {
	path   string
	errors []string
}

//
// An actionFunction takes a path prefix and a list of arguments,
// and contains whatever code will be ran as part of an action.
//
type actionFunction func([]string, *ArtifactWriter)

//
// All tasks in Dexter are stored here, added using the `add` function
// in the files that define the tasks.
//
var Tasks = map[string]Task{}

//
// Add a task to the Tasks package variable,
// using the task's name as the key.
//
func add(t Task) {
	if _, ok := Tasks[t.Name]; ok {
		log.WithFields(log.Fields{
			"at":   "tasks.add",
			"name": t.Name,
		}).Warn("task name already defined, overriding")
	}
	Tasks[t.Name] = t
}

//
// Run the Task's actionFunction unless Dexter
// isn't running on a platform the task supports.
//
func (task *Task) Run(dir string, args []string) {
	if len(task.supportedPlatforms) > 0 && !util.StringsInclude(task.supportedPlatforms, runtime.GOOS) {
		log.WithFields(log.Fields{
			"at":       "task.Run",
			"task":     task.Name,
			"platform": runtime.GOOS,
		}).Error("task not support on platform")
		return
	}
	writer := ArtifactWriter{
		path: dir + task.Name + "/",
	}
	task.actionFunction(
		args,
		&writer,
	)
	writer.flushErrors()
}

//
// Write a file to the filesystem, logging any errors
//
func (writer *ArtifactWriter) Write(dst string, data []byte) {
	dst = writer.path + dst
	dir := path.Dir(dst)
	err := os.MkdirAll(filepath.FromSlash(dir), 0700)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "tasks.Write",
			"path":  dir,
			"error": err.Error(),
		}).Error("unable to create directory for evidence")
		return
	}
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "tasks.Write",
			"file":  dst,
			"error": err.Error(),
		}).Error("unable to write piece of evidence for report")
	}
}

//
// Write an error into a tasks's report
//
func (writer *ArtifactWriter) Error(message string) {
	writer.errors = append(writer.errors, message)
}

//
// Write a task's errors to disk
//
func (writer *ArtifactWriter) flushErrors() {
	if len(writer.errors) > 0 {
		data := []byte{}
		for _, errstr := range writer.errors {
			data = append(data, []byte(errstr)...)
			data = append(data, []byte("\n")...)
		}
		data = append(data, []byte("\n")...)
		ioutil.WriteFile(writer.path+"errors.txt", data, 0644)
	}
}
