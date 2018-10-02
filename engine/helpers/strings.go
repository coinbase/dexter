package helpers

import (
	"fmt"

	"github.com/satori/go.uuid"
)

//
// Return a new randomly generated ID for a Dexter investigation
//
func NewDexterID() string {
	return uuid.Must(uuid.NewV4()).String()[0:8]
}

//
// For each task in a Dexter investigation, create a printable string
// version.  Return the slice of all of these.
//
func TaskStrings(tasks map[string][]string) []string {
	set := []string{}
	for key, value := range tasks {
		set = append(set, StringWithArgs(key, value, false))
	}
	return set
}

//
// Create a printable representation of a string with arguments,
// redacting the arguments if the private argument is true.
//
func StringWithArgs(item string, args []string, private bool) string {
	if len(args) == 0 {
		return item
	}
	argstr := ""
	for i, arg := range args {
		if private {
			argstr += "REDACTED"
		} else {
			argstr += "\"" + arg + "\""
		}
		if i != len(args)-1 {
			argstr += ", "
		}
	}
	return fmt.Sprintf("%s(%s)", item, argstr)
}
