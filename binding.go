/*
Package enflag provides functionality for parsing values from environment variables
and command-line flags, two common configuration sources for server-side applications.

Configuration values are prioritized in the following order:
flag > environment variable > default value. Both environment variables and flags are optional.

Flag parsing is handled by the standard library's flag package via the CommandLine flag set.
Like the flag package, errors encountered during environment variable parsing will cause
the program to exit with status code 2.

# Example usage:

	var port int
	Var(&port).Bind("PORT", "port")

	var ts time.Time
	Var(&ts).
	    WithFlagUsage("").
	    WithTimeLayout(time.DateOnly).
	    Bind("START_TIME", "start-time")

After all flags are defined, call

	enflag.Parse()
*/

package enflag

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atelpis/enflag/internal/parsers"
)

type builtin interface {
	[]byte |
		string | []string |
		int | []int | int64 | []int64 |
		uint | []uint | uint64 | []uint64 |
		float64 | []float64 |
		bool | []bool |
		time.Time | *time.Time | []time.Time |
		time.Duration | []time.Duration |
		url.URL | *url.URL | []url.URL |
		net.IP | *net.IP | []net.IP
}

// SliceSeparator is the default separator for parsing slices.
var SliceSeparator = ","

// TimeLayout is the default layout for parsing time.
var TimeLayout = time.RFC3339

// StringDecodeFunc is the default string-to-[]byte decoder.
var StringDecodeFunc = base64.StdEncoding.DecodeString

// Binding holds a pointer to a specified variable along with settings
// for parsing environment variables and command-line flags into it.
// It is a generic type constrained by `builtin`.
// For details on the supported types, refer to the `builtin` constraint.
//
// A Binding should always be created using the Var function and finalized
// by calling Bind(), BindEnv(), or BindFlag().
//
// Example usage:
//
//	var port int
//	Var(&port).Bind("PORT", "port")
type Binding[T builtin] struct {
	binding

	p   *T
	def T
}

// Var creates a new Binding for the given pointer p.
//
// The created Binding should be finalized by calling Bind(), BindEnv(),
// or BindFlag().
//
// Example usage:
//
//	var port int
//	Var(&port).Bind("PORT", "port")
//
// For more advanced usage, methods like WithDefault and WithSliceSeparator
// can be chained. For example:
//
//	var ts time.Time
//	Var(&ts).
//	    WithFlagUsage("").
//	    WithTimeLayout(time.DateOnly).
//	    Bind("START_TIME", "start-time")
func Var[T builtin](p *T) *Binding[T] {
	b := &Binding[T]{
		p: p,
	}
	b.sliceSep = SliceSeparator
	b.timeLayout = TimeLayout
	b.decoder = StringDecodeFunc

	return b
}

// WithDefault sets the default value for Binding.
func (b *Binding[T]) WithDefault(val T) *Binding[T] {
	b.def = val
	return b
}

// WithFlagUsage sets the help message for the bound command-line flag.
func (b *Binding[T]) WithFlagUsage(usage string) *Binding[T] {
	b.flagUsage = usage
	return b
}

// WithSliceSeparator sets a slice separator for the Binding.
// This is only applicable to slice types of the builtin constraint.
//
// If not explicitly set, the global variable SliceSeparator will be used.
// The default value of the SliceSeparator is ",".
func (b *Binding[T]) WithSliceSeparator(sep string) *Binding[T] {
	b.sliceSep = sep
	return b
}

// WithStringDecodeFunc sets a function for decoding a string into []byte.
// This is only applicable to []byte variables.
//
// If not explicitly set, the global variable StringDecodeFunc() will be used.
// The default decoder is base64.StdEncoding.DecodeString.
func (b *Binding[T]) WithStringDecodeFunc(f func(string) ([]byte, error)) *Binding[T] {
	b.decoder = f
	return b
}

// WithTimeLayout sets a layout for parsing time for this Binding.
// This is only applicable to time variables.
//
// If not explicitly set, the global variable TimeLayout() will be used.
// The default layout is time.RFC3339.
func (b *Binding[T]) WithTimeLayout(layout string) *Binding[T] {
	b.timeLayout = layout
	return b
}

// Bind registers an environment variable and a command-line flag
// as data sources for this Binding. Both sources are optional.
// Use BindEnv or BindFlag to bind a single source.
//
// Data sources are prioritized as follows:
// flag > environment variable > default value.
//
// If a flag is used, Parse() must be called after all bindings
// are created.
func (b *Binding[T]) Bind(envName string, flagName string) {
	b.envName, b.flagName = envName, flagName
	*b.p = b.def

	switch ptr := any(b.p).(type) {
	case *[]byte:
		handleVar(b.binding, ptr, b.decoder)

	case *string:
		handleVar(b.binding, ptr, parsers.String)

	case *[]string:
		handleSlice(b.binding, ptr, parsers.String)

	case *int:
		handleVar(b.binding, ptr, strconv.Atoi)

	case *[]int:
		handleSlice(b.binding, ptr, strconv.Atoi)

	case *int64:
		handleVar(b.binding, ptr, parsers.Inte64)

	case *[]int64:
		handleSlice(b.binding, ptr, parsers.Inte64)

	case *uint:
		handleVar(b.binding, ptr, parsers.Uint)

	case *[]uint:
		handleSlice(b.binding, ptr, parsers.Uint)

	case *uint64:
		handleVar(b.binding, ptr, parsers.Uint64)

	case *[]uint64:
		handleSlice(b.binding, ptr, parsers.Uint64)

	case *float64:
		handleVar(b.binding, ptr, parsers.Float64)

	case *[]float64:
		handleSlice(b.binding, ptr, parsers.Float64)

	case *bool:
		handleVar(b.binding, ptr, strconv.ParseBool)

	case *[]bool:
		handleSlice(b.binding, ptr, strconv.ParseBool)

	case *time.Time:
		handleVar(b.binding, ptr, parsers.Time(b.timeLayout))

	case **time.Time:
		handleVar(b.binding, ptr, parsers.Ptr(parsers.Time(b.timeLayout)))

	case *[]time.Time:
		handleSlice(b.binding, ptr, parsers.Time(b.timeLayout))

	case *time.Duration:
		handleVar(b.binding, ptr, time.ParseDuration)

	case *[]time.Duration:
		handleSlice(b.binding, ptr, time.ParseDuration)

	case *url.URL:
		handleVar(b.binding, ptr, parsers.URL)

	case **url.URL:
		handleVar(b.binding, ptr, url.Parse)

	case *[]url.URL:
		handleSlice(b.binding, ptr, parsers.URL)

	case *net.IP:
		handleVar(b.binding, ptr, parsers.IP)

	case **net.IP:
		handleVar(b.binding, ptr, parsers.Ptr(parsers.IP))

	case *[]net.IP:
		handleSlice(b.binding, ptr, parsers.IP)
	}
}

// BindEnv is a shorthand for Bind when only an environment variable is needed.
func (b *Binding[T]) BindEnv(name string) {
	b.Bind(name, "")
}

// BindFlag is a shorthand for Bind when only a command-line flag is needed.
func (b *Binding[T]) BindFlag(name string) {
	b.Bind("", name)
}

// CustomBinding holds a pointer to a variable along with a custom parser
// and additional settings.
//
// A CustomBinding should always be created using VarFunc or its alternatives,
// such as VarJSON, and must be finalized by calling Bind(), BindEnv(),
// or BindFlag().
type CustomBinding[T any] struct {
	binding

	p      *T
	def    T
	parser func(string) (T, error)
}

// VarFunc creates a new CustomBinding for the given pointer p and
// the specified string parser function. The parser function is used
// to convert a string into the desired type T and will be used to parse
// both the environment variable and the flag.
func VarFunc[T any](p *T, parser func(string) (T, error)) *CustomBinding[T] {
	b := CustomBinding[T]{
		p:      p,
		parser: parser,
	}

	return &b
}

// VarJSON creates a new CustomBinding for the given pointer p and
// uses JSON unmarshaling as the parser for both the environment variable
// and the flag.
func VarJSON[T any](p *T) *CustomBinding[T] {
	return VarFunc(p, func(s string) (T, error) {
		var d T
		err := json.Unmarshal([]byte(s), &d)
		return d, err
	})
}

// WithDefault sets the default value for the CustomBinding.
func (b *CustomBinding[T]) WithDefault(val T) *CustomBinding[T] {
	b.def = val
	return b
}

// WithFlagUsage sets the help message for the bound command-line flag.
func (b *CustomBinding[T]) WithFlagUsage(usage string) *CustomBinding[T] {
	b.flagUsage = usage
	return b
}

// Bind registers an environment variable and a command-line flag
// as data sources for this Binding. Both sources are optional.
// Use BindEnv or BindFlag to bind a single source.
//
// Data sources are prioritized as follows:
// flag > environment variable > default value.
//
// If a flag is used, Parse() must be called after all bindings
// are created.
func (b *CustomBinding[T]) Bind(envName string, flagName string) {
	b.envName, b.flagName = envName, flagName
	*b.p = b.def

	handleVar(b.binding, b.p, b.parser)

}

// BindEnv is a shorthand for Bind when only an environment variable is needed.
func (b *CustomBinding[T]) BindEnv(name string) {
	b.Bind(name, "")
}

// BindFlag is a shorthand for Bind when only a command-line flag is needed.
func (b *CustomBinding[T]) BindFlag(name string) {
	b.Bind("", name)
}

// Parse calls the standard library's `flag` package's `Parse()` function.
// Like the standard library's `flag` package, Parse() must be called
// after all flags have been defined.
func Parse() {
	flag.Parse()
}

type binding struct {
	envName   string
	flagName  string
	flagUsage string

	sliceSep   string
	decoder    func(string) ([]byte, error)
	timeLayout string
}

func handleVar[T any](b binding, ptr *T, parser func(string) (T, error)) {
	if envVal := os.Getenv(b.envName); envVal != "" {
		v, err := parser(envVal)
		if err != nil {
			fmt.Fprintf(
				flag.CommandLine.Output(),
				"Unable to parse env-variable %s as type %T\n",
				b.envName,
				*ptr,
			)

			// os.Exit(2) replicates the default error handling behavior of flag.CommandLine
			if !isTestEnv {
				os.Exit(2)
			}
		}
		*ptr = v
	}

	if b.flagName != "" {
		flag.Func(b.flagName, b.flagUsage, func(s string) error {
			parsed, err := parser(s)
			if err != nil {
				return err
			}

			*ptr = parsed
			return nil
		})
	}
}

func handleSlice[T any](b binding, ptr *[]T, parser func(string) (T, error)) {
	if envVal := os.Getenv(b.envName); envVal != "" {
		for _, v := range strings.Split(envVal, b.sliceSep) {
			parsed, err := parser(v)
			if err != nil {
				fmt.Fprintf(
					flag.CommandLine.Output(),
					"Unable to parse env-variable %s as type %T\n",
					b.envName,
					*ptr,
				)

				// os.Exit(2) replicates the default error handling behavior of flag.CommandLine
				if !isTestEnv {
					os.Exit(2)
				}
			}
			*ptr = append(*ptr, parsed)
		}
	}

	if b.flagName != "" {
		flag.Func(b.flagName, b.flagUsage, func(s string) error {
			for _, v := range strings.Split(s, b.sliceSep) {
				parsed, err := parser(v)
				if err != nil {
					return err
				}

				*ptr = append(*ptr, parsed)
			}

			return nil
		})
	}
}

var isTestEnv bool
