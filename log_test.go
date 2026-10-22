package selfupdate_test

import (
	stdlog "log"
	"os"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/github"
)

func ExampleConfig_logger() {
	// any logger providing the two methods Print and Printf can be used: the standard log.Logger does
	logger := stdlog.New(os.Stdout, "selfupdate ", 0)
	// the source and the updater each have their own Logger: give them the same one to see everything
	source, err := github.NewSource(github.Config{Logger: logger})
	if err != nil {
		panic(err)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source, Logger: logger})
	if err != nil {
		panic(err)
	}
	_ = updater
}

type testLogger struct {
	t *testing.T
}

// newTestLogger returns a Logger sending the messages to the test log.
func newTestLogger(t *testing.T) *testLogger {
	return &testLogger{t: t}
}

func (l *testLogger) Print(v ...any) {
	l.t.Log(v...)
}

func (l *testLogger) Printf(format string, v ...any) {
	l.t.Logf(format, v...)
}
