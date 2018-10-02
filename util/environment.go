package util

//
// All values for runtime.GOOS, derived from
// https://github.com/golang/go/blob/master/src/go/build/syslist.go
//
var AllPlatforms = []string{
	"android", "darwin", "dragonfly", "freebsd",
	"js", "linux", "nacl", "netbsd", "openbsd",
	"plan9", "solaris", "windows", "zos",
}

//
// A values for runtime.GOOS that are unix-like
//
var UnixLike = []string{
	"darwin", "dragonfly", "freebsd",
	"linux", "nacl", "netbsd", "openbsd",
	"plan9", "solaris",
}
