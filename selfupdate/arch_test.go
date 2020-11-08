package selfupdate

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOSAndArch(t *testing.T) {
	os, arch := GetOSArch()
	if os != runtime.GOOS {
		t.Errorf("OS should be %s but found %s", runtime.GOOS, os)
	}
	if arch[0] != runtime.GOARCH {
		t.Errorf("Arch should be %s but found %s", runtime.GOARCH, arch)
	}
}

func TestAdditionalArch(t *testing.T) {
	testData := []struct {
		arch     string
		goarm    uint8
		expected []string
	}{
		{"arm64", 8, []string{}},
		{"arm", 8, []string{}}, // armv8 is called arm64 - this shouldn't happen
		{"arm", 7, []string{"armv7", "armv6", "armv5"}},
		{"arm", 6, []string{"armv6", "armv5"}},
		{"arm", 5, []string{"armv5"}},
		{"arm", 4, []string{}}, // go is not supporting below armv5
	}

	for _, testItem := range testData {
		t.Run(fmt.Sprintf("%s-%d", testItem.arch, testItem.goarm), func(t *testing.T) {
			result := generateAdditionalArch(testItem.arch, testItem.goarm)
			assert.ElementsMatch(t, testItem.expected, result)
		})
	}
}
