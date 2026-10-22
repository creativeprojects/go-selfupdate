package selfupdate

// Logger receives the messages describing what the updater and the sources are doing.
// The interface is compatible with the standard library log.Logger.
type Logger interface {
	// Print calls Output to print to the standard logger. Arguments are handled in the manner of fmt.Print.
	Print(v ...any)
	// Printf calls Output to print to the standard logger. Arguments are handled in the manner of fmt.Printf.
	Printf(format string, v ...any)
}

// DiscardLogger drops every message.
// It is the default when no Logger is set in the Config of the updater or of a source.
var DiscardLogger Logger = discardLogger{}

type discardLogger struct{}

func (discardLogger) Print(v ...any)                 {}
func (discardLogger) Printf(format string, v ...any) {}
