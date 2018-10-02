package facts

//
// The init function should be used to declare your new fact when Dexter starts up.
// Just pass a Fact struct to the `add` function.
//
func init() {
	add(Fact{
		// Give your fact a short and descriptive name
		Name: "example-fact",

		// Your fact should also have a more detailed description that isn't too long
		Description: "this is an example fact, showing how facts work",

		// If your fact only makes sense in the context of some arguments, you can specify
		// a minimum number of arguments here.  Omit this if 0 arguments are ok.
		MinimumArguments: 1,

		// If this fact will be checking sensitive information that not all Dexter hosts should
		// be able to see, mark it as a private fact.  This will use a hash to hide the data
		// from the Dexter hosts.
		Private: true,

		// supportedPlatforms contains valid values for go's runtime.GOOS
		// If this is omitted, the default value is all platforms.
		//
		// The list of valid platforms is here: https://github.com/golang/go/blob/master/src/go/build/syslist.go
		//
		// Here we define unix-list systems
		supportedPlatforms: []string{
			"linux", "darwin", "dragonfly", "freebsd",
			"netbsd", "openbsd", "plan9", "solaris",
		},

		// Should an error be encountered while checking your fact, a default state can be set.
		// This is what the fact check will return in the case of an error.
		defaultState: true,

		// This is the actual function that contains the fact checking logic, defined below.
		function: exampleFact,
	})
}

//
// This is the code that will run to check if your fact is true.
//
// The arguments are the strings defined by the investigator when the investigation
// was generated.
//
// If the fact is private, the arguments will be hashed and salted, and need to be
// split before they can be checked.
//
// If this function returns an error, the fact's defaultState value will be used.
//
func exampleFact(args []string) (bool, error) {
	//
	// Here your fact can do what it is supposed to do (check if a file exists, check if
	// an envar exists, etc...).
	//

	//
	// If this is a private fact, each argument takes the form of <hash(argument) + salt>
	// The `splitDigestAndSalt` function must be used to separate them, then the argument
	// digest can be compared using the `Hash` function.
	//
	// For example, let's say arg[0] is a sensitive username.
	//
	// digest, salt := splitDigestAndSalt(arg[0])
	//
	// Now you want to check if the username "foo" is the argument to this investigation.
	// Create a hash of "foo", then compare the hashes.
	//
	// foo_hash := Hash("foo", salt)
	// return digest == foo_hash, nil
	//

	return false, nil
}
