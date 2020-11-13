package selfupdate

import (
	stdlog "log"
	"os"
)

func ExampleSetLogger() {
	// you can plug-in any logger providing the 2 methods Print and Printf
	// the default log.Logger satisfies the interface
	logger := stdlog.New(os.Stdout, "selfupdate ", 0)
	SetLogger(logger)
}
