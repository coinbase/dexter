package facts

import (
	"runtime"
)

func init() {
	add(Fact{
		Name:             "platform-is",
		Description:      "check if a host's runtime.GOOS platform matches a string",
		MinimumArguments: 1,
		function:         platformIs,
	})
}

func platformIs(args []string) (bool, error) {
	for _, arg := range args {
		if runtime.GOOS == arg {
			return true, nil
		}
	}
	return false, nil
}
