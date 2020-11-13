package selfupdate

import (
	_ "unsafe"
)

//go:linkname goarm runtime.goarm
var goarm uint8
