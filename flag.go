/*
Package enflag provides functionality for parsing values from environment variables
and command-line flags, two common configuration sources for server-side applications.

Configuration values are prioritized in the following order:
flag > environment variable > default value. Both environment variables and flags are optional.

enflag uses generics to provide two primary functions: Bind and BindFunc.

Flag parsing is handled by the standard library's flag package via the CommandLine flag set.
Like the flag package, errors encountered during environment variable parsing will cause
the program to exit with status code 2.

# Usage

var intval int
enflag.Bind(&intval, "ENV_NAME", "flag-name", 1234, "help message for flag-name")

After all flags are defined, call

	enflag.Parse()
*/
package enflag

import (
	"flag"
	"net"
	"net/url"
	"time"
)

// TODO: rename to Predefined or Builtin
type Bindable interface {
	[]byte |
		string | []string | bool | []bool |
		int | []int | int64 | []int64 | uint | []uint | uint64 | []uint64 | float64 | []float64 |
		time.Time | *time.Time | []time.Time | time.Duration | []time.Duration |
		url.URL | []url.URL | *url.URL | []*url.URL | net.IP | []net.IP
}

// Parse calls the standard library's `flag` package's `Parse()` function.
// Like the standard library's `flag` package, Parse() must be called
// after all flags have been defined.
func Parse() {
	flag.Parse()
}

var isTestEnv bool
