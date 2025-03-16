package enflag

import (
	"flag"
	"fmt"
	"os"
)

// ErrorHandlerFunc is a function called after a value parser returns an error.
// See predefined options: OnErrorExit, OnErrorIgnore, and OnErrorLogAndContinue.
// It can also be replaced with a custom handler.
var ErrorHandlerFunc = OnErrorExit

// OnErrorExit prints the error and exits with status code 2.
var OnErrorExit = func(err error, rawVal string, target any, envName string, flagName string) {
	OnErrorLogAndContinue(err, rawVal, target, envName, flagName)
	osExitFunc(2)
}

// OnErrorIgnore silently ignores the error.
// If a default value is specified, it will be used.
var OnErrorIgnore = func(err error, rawVal string, target any, envName string, flagName string) {}

// OnErrorLogAndContinue prints the error message but continues execution.
// If a default value is specified, it will be used.
var OnErrorLogAndContinue = func(err error, rawVal string, target any, envName string, flagName string) {
	_, _ = err, rawVal

	var msg string
	if envName != "" {
		msg = fmt.Sprintf("unable to parse env-variable %q as type %T\n", envName, target)
	} else if flagName != "" {
		msg = fmt.Sprintf("unable to parse flag %q as type %T\n", flagName, target)
	}

	flag.CommandLine.Output().Write([]byte(msg))
}

func handleError[T any](err error, target *T, rawVal, envName string, flagName string) {
	ErrorHandlerFunc(err, rawVal, *target, envName, flagName)
}

var osExitFunc = os.Exit
