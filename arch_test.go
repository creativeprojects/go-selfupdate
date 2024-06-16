package selfupdate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdditionalArch(t *testing.T) {
	testData := []struct {
		arch          string
		goarm         uint8
		universalArch string
		expected      []string
	}{
		{"arm64", 0, "", []string{"arm64"}},
		{"arm64", 0, "all", []string{"arm64", "all"}},
		{"arm", 8, "", []string{"arm"}}, // armv8 is called arm64 - this shouldn't happen
		{"arm", 7, "", []string{"armv7", "armv6", "armv5", "arm"}},
		{"arm", 6, "", []string{"armv6", "armv5", "arm"}},
		{"arm", 5, "", []string{"armv5", "arm"}},
		{"arm", 4, "", []string{"arm"}}, // go is not supporting below armv5
		{"amd64", 0, "", []string{"amd64", "x86_64"}},
		{"amd64", 0, "all", []string{"amd64", "x86_64", "all"}},
	}

	for _, testItem := range testData {
		t.Run(fmt.Sprintf("%s-%d", testItem.arch, testItem.goarm), func(t *testing.T) {
			result := getAdditionalArch(testItem.arch, testItem.goarm, testItem.universalArch)
			assert.Equal(t, testItem.expected, result)
		})
	}
}
