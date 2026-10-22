package update

import "os"

// Options for Apply update
type Options struct {
	// TargetPath defines the path to the file to update.
	// The empty string means 'the executable file of the running program'.
	// If the file does not exist, it is created.
	TargetPath string

	// Create TargetPath replacement with this file mode. If zero, defaults to 0755.
	TargetMode os.FileMode

	// Store the old executable file at this path after a successful update.
	// The empty string means the old executable file will be removed after the update.
	OldSavePath string
}
