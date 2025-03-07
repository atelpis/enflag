package enflag

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
)

// TODO: proper name
type StringDecodeFunc func(string) ([]byte, error)

// TODO: consider moving to a sub-package
var Base64DecodeFunc StringDecodeFunc = base64.StdEncoding.DecodeString
var HexDecodeFunc StringDecodeFunc = hex.DecodeString

func test() {

	var t int
	// NewBinder(&t).WithDefault(7).Bind("TEST", "test")
	Var(&t).BindEnv("TEST")

	// SLICES
	var st []int
	// NewBinder(&st).WithSliceSeparator(":").Bind("TEST-SL", "test-sl")
	Var(&st).WithSliceSeparator(":").Bind("TEST-SL", "test-sl")

	// SLICES FUNC
	var t2 int
	VarFunc(&t2, strconv.Atoi).WithSliceSeparator(";").Bind("TEST2", "test2")

	// BYTES
	var secret []byte
	Var(&secret).WithDecoder(HexDecodeFunc).Bind("e", "f")

	// Parse JSON string into the target
	type User struct {
		ID   int
		Name string
	}
	var usr User
	VarJSON(&usr).Bind("e2", "f")

	// TIME
	var ts time.Time
	Var(&ts).WithTimeFormat(time.RFC1123).Bind("e", "f")
}

type Binding[T Bindable] struct {
	binding

	p   *T
	def T

	// if the target is []byte, decoder will be used to decode the input string
	decoder    StringDecodeFunc
	timeFormat string
}

func Var[T Bindable](p *T) *Binding[T] {
	return &Binding[T]{
		binding: defaultBinding,
		p:       p,

		timeFormat: time.RFC3339,
		decoder:    Base64DecodeFunc,
	}
}

func VarFunc[T any](p *T, parser func(string) (T, error)) *CustomBinding[T] {
	b := CustomBinding[T]{
		binding: defaultBinding,
		p:       p,
		parser:  parser,
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
	Bind(b.p, envName, flagName, b.def, b.flagUsage)
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

func (b *CustomBinding[T]) WithSliceSeparator(sep string) *CustomBinding[T] {
	b.sliceSep = sep
	return b
}

func (b *CustomBinding[T]) Bind(envName string, flagName string) {
	BindFunc(b.p, envName, flagName, b.def, b.flagUsage, b.parser)
}

func (b *CustomBinding[T]) BindEnv(name string) {
	b.Bind(name, "")
}

func (b *CustomBinding[T]) BindFlag(name string) {
	b.Bind("", name)
}

type binding struct {
	flagUsage string
	sliceSep  string

	// Possible extras:
	// shortFlag string
	// valMutator func(string) string
}

var defaultBinding = binding{
	sliceSep: ",",
}
