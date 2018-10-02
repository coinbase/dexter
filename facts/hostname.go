package facts

import (
	"os"
	"strings"
)

func init() {
	add(Fact{
		Name:             "hostname-contains",
		Description:      "check if the host's hostname contains the argument as a substring",
		MinimumArguments: 1,
		function:         hostnameContains,
	})
	add(Fact{
		Name:             "hostname-is",
		Description:      "check if the host's hostname is an exact match to the argument",
		MinimumArguments: 1,
		function:         hostnameIs,
	})
}

func hostnameContains(args []string) (bool, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return false, err
	}
	for _, arg := range args {
		if strings.Contains(hostname, arg) {
			return true, nil
		}
	}
	return false, nil
}

func hostnameIs(args []string) (bool, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return false, err
	}
	for _, arg := range args {
		if hostname == arg {
			return true, nil
		}
	}
	return false, nil
}
