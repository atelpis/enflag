package enflag

import (
	"flag"
	"fmt"
	"os"
)

var ErrorHandlerFunc = OnErrorExit

var OnErrorExit = func(rawVal string, target any, envName string, flagName string) {
	OnErrorContinue(rawVal, target, envName, flagName)
	os.Exit(2)
}

var OnErrorIgnore = func(rawVal string, target any, envName string, flagName string) {}

var OnErrorContinue = func(rawVal string, target any, envName string, flagName string) {
	flag.CommandLine.Output().Write([]byte(errorMessage(rawVal, target, envName, flagName)))
}

// only one of envName or flagName should be set
func errorMessage(rawVal string, target any, envName string, flagName string) string {
	if envName != "" {
		return fmt.Sprintf("unable to parse env-variable %s as type %T\n", envName, target)
	}

	if flagName != "" {
		return fmt.Sprintf("unable to parse flag %s as type %T\n", flagName, target)
	}

	return ""
}

func handleError[T any](ptr *T, rawVal, envName string, flagName string) {
	ErrorHandlerFunc(rawVal, *ptr, envName, flagName)
}
