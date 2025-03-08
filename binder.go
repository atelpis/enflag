package enflag

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	flagPkg "flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: proper name
type StringDecodeFunc func(string) ([]byte, error)

// TODO: consider moving to a sub-package
var Base64DecodeFunc StringDecodeFunc = base64.StdEncoding.DecodeString
var HexDecodeFunc StringDecodeFunc = hex.DecodeString

type Binding[T Builtin] struct {
	binding

	p   *T
	def T
}

func Var[T Builtin](p *T) *Binding[T] {
	b := &Binding[T]{
		p: p,
	}
	b.timeFormat = time.RFC3339
	b.decoder = Base64DecodeFunc
	b.sliceSep = ","

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

func (b *Binding[T]) WithDecoder(f func(string) ([]byte, error)) *Binding[T] {
	b.decoder = f
	return b
}

func (b *Binding[T]) WithTimeFormat(layout string) *Binding[T] {
	b.timeFormat = layout
	return b
}

func (b *Binding[T]) Bind(envName string, flagName string) {
	b.envName, b.flagName = envName, flagName
	*b.p = b.def

	switch ptr := any(b.p).(type) {
	case *[]byte:
		handleVar(b.binding, ptr, b.decoder, nil)

	case *string:
		handleVar(
			b.binding,
			ptr,
			func(s string) (string, error) {
				return s, nil
			},
			flagPkg.StringVar,
		)

	case *[]string:
		// TODO:

	case *int:
		handleVar(
			b.binding,
			ptr,
			strconv.Atoi,
			flagPkg.IntVar,
		)

	case *[]int:
		handleSlice(b.binding, ptr, strconv.Atoi)

	case *int64:
		handleVar(
			b.binding,
			ptr,
			func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			flagPkg.Int64Var,
		)

	case *[]int64:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
		)

	case *uint:
		handleVar(
			b.binding,
			ptr,
			func(s string) (uint, error) {
				v, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					return 0, err
				}
				return uint(v), nil
			},
			flagPkg.UintVar,
		)

	case *[]uint:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (uint, error) {
				v, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					return 0, err
				}
				return uint(v), nil
			},
		)

	case *uint64:
		handleVar(
			b.binding,
			ptr,
			func(s string) (uint64, error) {
				return strconv.ParseUint(s, 10, 64)
			},
			flagPkg.Uint64Var,
		)

	case *[]uint64:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (uint64, error) {
				return strconv.ParseUint(s, 10, 64)
			},
		)

	case *float64:
		handleVar(
			b.binding,
			ptr,
			func(s string) (float64, error) {
				return strconv.ParseFloat(s, 10)
			},
			flagPkg.Float64Var,
		)

	case *[]float64:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (float64, error) {
				return strconv.ParseFloat(s, 10)
			},
		)

	case *bool:
		handleVar(b.binding, ptr, strconv.ParseBool, flagPkg.BoolVar)

	case *[]bool:
		handleSlice(b.binding, ptr, strconv.ParseBool)

	case *time.Time:
		handleVar(
			b.binding,
			ptr,
			func(s string) (time.Time, error) {
				return time.Parse(b.timeFormat, s)
			},
			nil,
		)

	case **time.Time:
		handleVar(
			b.binding,
			ptr,
			func(s string) (*time.Time, error) {
				ts, err := time.Parse(b.timeFormat, s)
				return &ts, err
			},
			nil,
		)

	case *[]time.Time:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (time.Time, error) {
				return time.Parse(b.timeFormat, s)
			},
		)

	case *time.Duration:
		handleVar(
			b.binding,
			ptr,
			time.ParseDuration,
			flagPkg.DurationVar,
		)

	case *[]time.Duration:
		handleSlice(
			b.binding,
			ptr,
			time.ParseDuration,
		)

	case *url.URL:
		handleVar(
			b.binding,
			ptr,
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
		handleVar(
			b.binding,
			ptr,
			func(s string) (*url.URL, error) { return url.Parse(s) },
			nil,
		)

	case *[]url.URL:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (url.URL, error) {
				u, err := url.Parse(s)
				if err != nil {
					return url.URL{}, err
				}
				return *u, nil
			},
		)

	case *net.IP:
		handleVar(
			b.binding,
			ptr,
			func(s string) (net.IP, error) {
				ip := net.ParseIP(s)
				if ip == nil {
					return nil, errors.New("invalid IP address")
				}
				return ip, nil
			},
			nil,
		)

	case **net.IP:
		handleVar(
			b.binding,
			ptr,
			func(s string) (*net.IP, error) {
				ip := net.ParseIP(s)
				if ip == nil {
					return nil, errors.New("invalid IP address")
				}
				return &ip, nil
			},
			nil,
		)

	case *[]net.IP:
		handleSlice(
			b.binding,
			ptr,
			func(s string) (net.IP, error) {
				ip := net.ParseIP(s)
				if ip == nil {
					return nil, errors.New("invalid IP address")
				}
				return ip, nil
			},
		)
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

	handleVar(
		b.binding,
		b.p,
		b.parser,
		nil,
	)

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

	// Bindable-specific fields
	// if the target is []byte, decoder will be used to decode the input string
	sliceSep   string
	decoder    StringDecodeFunc
	timeFormat string
}

func handleVar[T any](
	b binding,
	ptr *T,
	parser func(string) (T, error),
	stdFlagFunc func(*T, string, T, string),
) {
	if envVal := os.Getenv(b.envName); envVal != "" {
		v, err := parser(envVal)
		if err != nil {
			fmt.Fprintf(
				flagPkg.CommandLine.Output(),
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
		if stdFlagFunc != nil {
			stdFlagFunc(ptr, b.flagName, *ptr, b.flagUsage)
		} else {
			flagPkg.Func(b.flagName, b.flagUsage, func(s string) error {
				parsed, err := parser(s)
				if err != nil {
					return err
				}

				*ptr = parsed
				return nil
			})
		}
	}
}

func handleSlice[T any](
	b binding,
	ptr *[]T,
	parser func(string) (T, error),
) {
	if envVal := os.Getenv(b.envName); envVal != "" {
		for _, v := range strings.Split(envVal, b.sliceSep) {
			parsed, err := parser(v)
			if err != nil {
				fmt.Fprintf(
					flagPkg.CommandLine.Output(),
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
		flagPkg.Func(b.flagName, b.flagUsage, func(s string) error {
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
