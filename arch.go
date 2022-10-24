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
	if arch == "arm" && goarm >= minARM && goarm <= maxARM {
		additionalArch := make([]string, 0, maxARM-minARM)
		for v := goarm; v >= minARM; v-- {
			additionalArch = append(additionalArch, fmt.Sprintf("armv%d", v))
		}
		return additionalArch
	}
	if arch == "amd64" {
		return []string{"x86_64"}
	}
	return []string{}
}
