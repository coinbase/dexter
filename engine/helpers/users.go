package helpers

import (
	"errors"
	"os/exec"
	"strings"
)

var stubbedLocalUsers = []string{}

//
// Stub all calls to LocalUsers with a string slice.
// Useful for testing.
//
func StubLocalUsers(set []string) {
	stubbedLocalUsers = set
}

//
// Look up local user names from /etc/passwd
//
func LocalUsers() ([]string, error) {
	if len(stubbedLocalUsers) != 0 {
		return stubbedLocalUsers, nil
	}

	stdOut, err := exec.Command("cut", "-d:", "-f1", "/etc/passwd").Output()
	if err != nil {
		return []string{}, err
	}
	names := strings.Split(string(stdOut), "\n")
	if len(names) == 0 {
		return []string{}, errors.New("command returned no names from /etc/passwd")
	}

	return names, nil
}
