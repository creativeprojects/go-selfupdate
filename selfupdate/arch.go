package selfupdate

import (
	"fmt"
	"runtime"
)

const (
	minARM = 5
	maxARM = 7
)

var (
	runtimeOS      = runtime.GOOS
	runtimeArch    = runtime.GOARCH
	additionalArch = generateAdditionalArch(runtimeArch, goarm)
)

// generateAdditionalArch we can use depending on the type of CPU
func generateAdditionalArch(arch string, goarm uint8) []string {
	additionalArch := make([]string, 0, maxARM-minARM)
	if arch == "arm" && goarm >= minARM && goarm <= maxARM {
		for v := goarm; v >= minARM; v-- {
			additionalArch = append(additionalArch, fmt.Sprintf("armv%d", v))
		}
	}
	return additionalArch
}

// GetOSArch returns the OS and Architecture(s) currently used to detect a new version
func GetOSArch() (os string, arch []string) {
	return runtimeOS, append([]string{runtimeArch}, additionalArch...)
}
