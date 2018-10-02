package facts

import (
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/util"
)

func init() {
	add(Fact{
		Name:               "user-exists",
		Description:        "check if a named user exists on the system",
		MinimumArguments:   1,
		Private:            true,
		supportedPlatforms: util.UnixLike,
		function:           userExists,
		defaultState:       true,
	})
}

func userExists(hashed_args []string) (bool, error) {
	names, err := helpers.LocalUsers()
	if err != nil {
		return false, err
	}

	for _, name := range names {
		for _, hashed_arg := range hashed_args {
			digest, salt := splitDigestAndSalt(hashed_arg)
			if digest == Hash(name, salt) {
				return true, nil
			}
		}
	}
	return false, nil
}
