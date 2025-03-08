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

var SliceSeparator = ","
var TimeLayout = time.RFC3339
var StringDecodeFunc = base64.StdEncoding.DecodeString

type Binding[T builtin] struct {
	binding

	p   *T
	def T
}

func Var[T builtin](p *T) *Binding[T] {
	b := &Binding[T]{
		p: p,
	}
	b.sliceSep = SliceSeparator
	b.timeLayout = TimeLayout
	b.decoder = StringDecodeFunc

	return b
}

func VarFunc[T any](p *T, parser func(string) (T, error)) *CustomBinding[T] {
	b := CustomBinding[T]{
		p:      p,
		parser: parser,
	}

	return &b
}

func VarJSON[T any](p *T) *CustomBinding[T] {
	return VarFunc(p, func(s string) (T, error) {
		var d T
		err := json.Unmarshal([]byte(s), &d)
		return d, err
	})
}

func (b *Binding[T]) WithDefault(val T) *Binding[T] {
	b.def = val
	return b
}

func (b *Binding[T]) WithFlagUsage(usage string) *Binding[T] {
	b.flagUsage = usage
	return b
}

func (b *Binding[T]) WithSliceSeparator(sep string) *Binding[T] {
	b.sliceSep = sep
	return b
}

func (b *Binding[T]) WithStringDecodeFunc(f func(string) ([]byte, error)) *Binding[T] {
	b.decoder = f
	return b
}

func (b *Binding[T]) WithTimeLayout(layout string) *Binding[T] {
	b.timeLayout = layout
	return b
}

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

func (b *Binding[T]) BindEnv(name string) {
	b.Bind(name, "")
}

func (b *Binding[T]) BindFlag(name string) {
	b.Bind("", name)
}

type CustomBinding[T any] struct {
	binding

	p      *T
	def    T
	parser func(string) (T, error)
}

func (b *CustomBinding[T]) WithDefault(val T) *CustomBinding[T] {
	b.def = val
	return b
}

func (b *CustomBinding[T]) WithFlagUsage(usage string) *CustomBinding[T] {
	b.flagUsage = usage
	return b
}

func (b *CustomBinding[T]) Bind(envName string, flagName string) {
	b.envName, b.flagName = envName, flagName
	*b.p = b.def

	handleVar(b.binding, b.p, b.parser)

}

func (b *CustomBinding[T]) BindEnv(name string) {
	b.Bind(name, "")
}

func (b *CustomBinding[T]) BindFlag(name string) {
	b.Bind("", name)
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
