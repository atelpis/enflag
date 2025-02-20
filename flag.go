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
	"errors"
	flagPkg "flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Bindable interface {
	string | int | int64 | uint | uint64 | float64 | bool | time.Duration | url.URL | *url.URL | net.IP
}

// Bind assigns a value to the pointer p, prioritizing sources in the following order:
// flag > environment variable > default value.
// Both the environment variable and flag sources are optional.
// Flag parsing and error handling are handled by the standard library's `flag` package
// using the default settings for the CommandLine flag set.
// An error during environment variable parsing will cause the program to exit with status code 2.
// Bind accepts values of Bindable types. For other types, use BindFunc.
func Bind[T Bindable](p *T, env string, flag string, value T, usage string) {
	*p = value

	switch ptr := any(p).(type) {
	case *string:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (string, error) {
				return s, nil
			},
			flagPkg.StringVar,
		)

	case *int:
		handle(
			ptr,
			env,
			flag,
			usage,
			strconv.Atoi,
			flagPkg.IntVar,
		)

	case *int64:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			flagPkg.Int64Var,
		)

	case *uint:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (uint, error) {
				v, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					return 0, err
				}
				return uint(v), nil
			},
			flagPkg.UintVar,
		)

	case *uint64:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (uint64, error) {
				return strconv.ParseUint(s, 10, 64)
			},
			flagPkg.Uint64Var,
		)

	case *float64:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (float64, error) {
				return strconv.ParseFloat(s, 10)
			},
			flagPkg.Float64Var,
		)

	case *bool:
		handle(
			ptr,
			env,
			flag,
			usage,
			strconv.ParseBool,
			flagPkg.BoolVar,
		)

	case *time.Duration:
		handle(
			ptr,
			env,
			flag,
			usage,
			time.ParseDuration,
			flagPkg.DurationVar,
		)

	case *url.URL:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (url.URL, error) {
				u, err := url.Parse(s)
				if err != nil {
					return url.URL{}, err
				}
				return *u, nil
			},
			nil,
		)

	case **url.URL:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (*url.URL, error) { return url.Parse(s) },
			nil,
		)

	case *net.IP:
		handle(
			ptr,
			env,
			flag,
			usage,
			func(s string) (net.IP, error) {
				ip := net.ParseIP(s)
				if ip == nil {
					return nil, errors.New("invalid IP address")
				}
				return ip, nil
			},
			nil,
		)
	}
}

// BindFunc works like Bind(), but accepts a custom string parser,
// which is used to parse both the environment variable and the flag value.
func BindFunc[T any](p *T, env string, flag string, value T, usage string, parser func(s string) (T, error)) {
	*p = value

	handle(
		p,
		env,
		flag,
		usage,
		parser,
		nil,
	)
}

// Parse calls the standard library's `flag` package's `Parse()` function.
// Like the standard library's `flag` package, Parse() must be called
// after all flags have been defined.
func Parse() {
	flagPkg.Parse()
}

func handle[T any](
	p *T,
	env string,
	flag string,
	usage string,
	parser func(s string) (T, error),
	stdFlagFunc func(*T, string, T, string),
) {
	if envVal := os.Getenv(env); envVal != "" {
		v, err := parser(envVal)
		if err != nil {
			fmt.Fprintf(
				flagPkg.CommandLine.Output(),
				"Unable to parse env-variable %s as type %T\n",
				env,
				*p,
			)

			// os.Exit(2) replicates the default error handling behavior of flag.CommandLine
			if !isTestEnv {
				os.Exit(2)
			}
		}
		*p = v
	}

	if flag != "" {
		if stdFlagFunc != nil {
			stdFlagFunc(p, flag, *p, usage)
		} else {
			flagPkg.Func(flag, usage, func(s string) error {
				parsed, err := parser(s)
				if err != nil {
					return err
				}

				*p = parsed
				return nil
			})
		}
	}
}

var isTestEnv bool
