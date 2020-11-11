package selfupdate

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

// Updater is responsible for managing the context of self-update.
// It contains GitHub client and its context.
type Updater struct {
	api       *github.Client
	apiCtx    context.Context
	validator Validator
	filters   []*regexp.Regexp
	os        string
	arch      string
	arm       uint8
}

// NewUpdater creates a new updater instance. It initializes GitHub API client.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
func NewUpdater(config Config) (*Updater, error) {
	token := config.APIToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}
	hc := newHTTPClient(ctx, token)

	filtersRe := make([]*regexp.Regexp, 0, len(config.Filters))
	for _, filter := range config.Filters {
		re, err := regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("could not compile regular expression %q for filtering releases: %v", filter, err)
		}
		filtersRe = append(filtersRe, re)
	}

	os := config.OS
	arch := config.Arch
	if os == "" {
		os = runtime.GOOS
	}
	if arch == "" {
		arch = runtime.GOARCH
	}
	arm := config.Arm
	if arm == 0 && goarm > 0 {
		arm = goarm
	}

	if config.EnterpriseBaseURL == "" {
		client := github.NewClient(hc)
		return &Updater{
			api:       client,
			apiCtx:    ctx,
			validator: config.Validator,
			filters:   filtersRe,
			os:        os,
			arch:      arch,
			arm:       arm,
		}, nil
	}

	u := config.EnterpriseUploadURL
	if u == "" {
		u = config.EnterpriseBaseURL
	}
	client, err := github.NewEnterpriseClient(config.EnterpriseBaseURL, u, hc)
	if err != nil {
		return nil, err
	}
	return &Updater{
		api:       client,
		apiCtx:    ctx,
		validator: config.Validator,
		filters:   filtersRe,
		os:        os,
		arch:      arch,
		arm:       arm,
	}, nil
}

// DefaultUpdater creates a new updater instance with default configuration.
// It initializes GitHub API client with default API base URL.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
func DefaultUpdater() *Updater {
	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	client := newHTTPClient(ctx, token)
	return &Updater{
		api:    github.NewClient(client),
		apiCtx: ctx,
		os:     runtime.GOOS,
		arch:   runtime.GOARCH,
		arm:    goarm,
	}
}

func newHTTPClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}
