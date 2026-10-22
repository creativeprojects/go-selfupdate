package selfupdate

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/creativeprojects/go-selfupdate/v2/internal"
)

// Updater is responsible for managing the context of self-update.
type Updater struct {
	source        Source
	validator     Validator
	filters       []*regexp.Regexp
	os            string
	arch          string
	arm           uint8
	universalArch string // only filled in when needed
	prerelease    bool
	draft         bool
	oldSavePath   string
	logger        Logger
}

// NewUpdater creates a new updater instance.
// Returns an error if the source is not specified in the config.
func NewUpdater(config Config) (*Updater, error) {
	source := config.Source
	if source == nil {
		return nil, ErrInvalidSource
	}

	filtersRe := make([]*regexp.Regexp, 0, len(config.Filters))
	for _, filter := range config.Filters {
		re, err := regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("could not compile regular expression %q for filtering releases: %w", filter, err)
		}
		filtersRe = append(filtersRe, re)
	}

	os := config.OS
	if os == "" {
		os = runtime.GOOS
	}
	arch := config.Arch
	if arch == "" {
		arch = runtime.GOARCH
	}
	arm := config.Arm
	if arm == 0 && strings.EqualFold(arch, "arm") {
		exe, _ := internal.GetExecutablePath()
		arm = getGOARM(exe)
	}
	universalArch := ""
	if strings.EqualFold(os, "darwin") && config.UniversalArch != "" {
		universalArch = config.UniversalArch
	}
	logger := config.Logger
	if logger == nil {
		logger = DiscardLogger
	}

	return &Updater{
		source:        source,
		validator:     config.Validator,
		filters:       filtersRe,
		os:            os,
		arch:          arch,
		arm:           arm,
		universalArch: universalArch,
		prerelease:    config.Prerelease,
		draft:         config.Draft,
		oldSavePath:   config.OldSavePath,
		logger:        logger,
	}, nil
}
