package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGOARM(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows")
	}
	t.Parallel()

	testCases := []struct {
		goOS          string
		goArch        string
		goArm         string
		expectedGoArm uint8
	}{
		{"linux", "arm", "7", 7},
		{"linux", "arm", "6", 6},
		{"linux", "arm", "5", 5},
		{"linux", "arm", "", 7}, // armv7 is the default
		{"linux", "arm64", "", 0},
		{"linux", "amd64", "", 0},
		{"darwin", "arm64", "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.goOS+" "+tc.goArch+" "+tc.goArm, func(t *testing.T) {
			tempBinary := t.TempDir() + "/tempBinary-" + tc.goOS + tc.goArch + "v" + tc.goArm
			buildCmd := fmt.Sprintf("GOOS=%s GOARCH=%s GOARM=%s go build -o %s ./testdata/hello", tc.goOS, tc.goArch, tc.goArm, tempBinary)
			cmd := exec.Command("sh", "-c", buildCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			require.NoError(t, err)

			goArm := getGOARM(tempBinary)
			assert.Equal(t, tc.expectedGoArm, goArm)
		})
	}
}
