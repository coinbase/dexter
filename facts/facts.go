//
// This package contains the facts Dexter can use to check if a host
// is in scope for an investigation.
//
package facts

import (
	"encoding/hex"
	log "github.com/Sirupsen/logrus"
	"github.com/coinbase/dexter/util"
	"golang.org/x/crypto/argon2"
	"runtime"
)

type Fact struct {
	Name               string
	Description        string
	Private            bool
	MinimumArguments   int
	Salt               string
	supportedPlatforms []string
	function           func([]string) (bool, error)
	defaultState       bool
}

var Facts = map[string]Fact{}

func add(f Fact) {
	if _, ok := Facts[f.Name]; ok {
		log.WithFields(log.Fields{
			"at":   "facts.add",
			"name": f.Name,
		}).Warn("fact name already defined, overriding")
	}
	Facts[f.Name] = f
}

//
// Look up a fact by name, returning the fact and a boolean to
// confirm the fact exists
//
func Get(name string) (Fact, bool) {
	check, ok := Facts[name]
	return check, ok
}

//
// Check if this fact indicates this host is in scope
//
func (checker *Fact) Assert(args []string) bool {
	// Ensure this check can be ran on this platform
	if len(checker.supportedPlatforms) > 0 && !util.StringsInclude(checker.supportedPlatforms, runtime.GOOS) {
		log.WithFields(log.Fields{
			"at":            "facts.Assert",
			"fact":          checker.Name,
			"platform":      runtime.GOOS,
			"default_state": checker.defaultState,
		}).Error("fact not support on platform, returning default state")
		return checker.defaultState
	}

	// Make a copy of the arguments, adding the investigation ID salt
	// if needed.
	saltedArgs := make([]string, len(args))
	for i, arg := range args {
		if checker.Private {
			saltedArgs[i] = arg + checker.Salt
		} else {
			saltedArgs[i] = arg
		}
	}
	// Run the checking function, returning the result,
	// or default state if there is an error
	result, err := checker.function(saltedArgs)
	if err == nil {
		return result
	} else {
		log.WithFields(log.Fields{
			"at":            "facts.Assert",
			"fact":          checker.Name,
			"platform":      runtime.GOOS,
			"default_state": checker.defaultState,
			"error":         err.Error(),
		}).Error("error running fact assert function, returning default state")
	}
	return checker.defaultState
}

func Hash(value, salt string) string {
	return hex.EncodeToString(
		argon2.IDKey([]byte(value), []byte(salt), 1, 64*1024, 4, 32),
	)
}

func splitDigestAndSalt(data string) (string, string) {
	digest := data[:len(data)-8]
	salt := data[len(data)-8:]
	return digest, salt
}
