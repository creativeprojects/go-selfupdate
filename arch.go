package selfupdate

import (
	"fmt"
)

const (
	minARM = 5
	maxARM = 7
)

// generateAdditionalArch we can use depending on the type of CPU
func generateAdditionalArch(arch string, goarm uint8) []string {
	additionalArch := make([]string, 0, maxARM-minARM)
	if arch == "arm" && goarm >= minARM && goarm <= maxARM {
		for v := goarm; v >= minARM; v-- {
			additionalArch = append(additionalArch, fmt.Sprintf("armv%d", v))
		}
	}
	if arch == "amd64" {
		additionalArch = append(additionalArch, "x86_64")
	}
	return additionalArch
}
