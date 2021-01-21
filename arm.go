package selfupdate

import (
	// unsafe is used to get a private variable from the runtime package
	_ "unsafe"
)

//go:linkname goarm runtime.goarm
var goarm uint8
