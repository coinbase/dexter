package facts

import (
	"github.com/coinbase/dexter/engine/helpers"
	"strings"
)

func init() {
	add(Fact{
		Name:             "project-name-contains",
		Description:      "check if the host's project name configuration contains the argument as a substring",
		MinimumArguments: 1,
		function:         projectNameContains,
	})
	add(Fact{
		Name:             "project-name-is",
		Description:      "check if the host's project name configuration is an exact match to the argument",
		MinimumArguments: 1,
		function:         projectNameIs,
	})
}

func projectNameContains(args []string) (bool, error) {
	for _, arg := range args {
		if strings.Contains(helpers.ProjectName(), arg) {
			return true, nil
		}
	}
	return false, nil
}

func projectNameIs(args []string) (bool, error) {
	for _, arg := range args {
		if helpers.ProjectName() == arg {
			return true, nil
		}
	}
	return false, nil
}
