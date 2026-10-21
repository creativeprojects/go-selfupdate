package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/gitea"
	"github.com/creativeprojects/go-selfupdate/github"
	"github.com/creativeprojects/go-selfupdate/gitlab"
	"github.com/creativeprojects/go-selfupdate/http"
)

// SplitDomainSlug tries to make sense of the repository string
// and returns a domain name (if present) and a slug.
//
// Example of valid entries:
//
//   - "owner/name"
//   - "github.com/owner/name"
//   - "http://github.com/owner/name"
func SplitDomainSlug(repo string) (domain, slug string, err error) {
	// simple case first => only a slug
	parts := strings.Split(repo, "/")
	if len(parts) == 2 {
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid slug or URL %q", repo)
		}
		return "", repo, nil
	}
	// trim trailing /
	repo = strings.TrimSuffix(repo, "/")

	if !strings.HasPrefix(repo, "http") && !strings.Contains(repo, "://") && !strings.HasPrefix(repo, "/") {
		// add missing scheme
		repo = "https://" + repo
	}

	repoURL, err := url.Parse(repo)
	if err != nil {
		return "", "", err
	}

	// make sure hostname looks like a real domain name
	if !strings.Contains(repoURL.Hostname(), ".") {
		return "", "", fmt.Errorf("invalid domain name %q", repoURL.Hostname())
	}
	domain = repoURL.Scheme + "://" + repoURL.Host
	slug = strings.TrimPrefix(repoURL.Path, "/")

	if slug == "" {
		return "", "", fmt.Errorf("invalid URL %q", repo)
	}
	return domain, slug, nil
}

func GetSource(ctx context.Context, cvsType, domain string) (selfupdate.Source, error) {
	if cvsType != "auto" && cvsType != "" {
		source, err := getSourceFromName(ctx, cvsType, domain)
		if err != nil {
			return nil, err
		}
		return source, nil
	}

	source, err := getSourceFromURL(ctx, domain)
	if err != nil {
		return nil, err
	}
	return source, nil
}

func getSourceFromName(ctx context.Context, name, domain string) (selfupdate.Source, error) {
	switch name {
	case "gitea":
		return gitea.NewSource(ctx, gitea.Config{BaseURL: domain})

	case "gitlab":
		return gitlab.NewSource(gitlab.Config{BaseURL: domain})

	case "http":
		return http.NewSource(http.Config{BaseURL: domain})

	default:
		return newGitHubSource(domain)
	}
}

func getSourceFromURL(ctx context.Context, domain string) (selfupdate.Source, error) {
	if strings.Contains(domain, "gitea") {
		return gitea.NewSource(ctx, gitea.Config{BaseURL: domain})
	}
	if strings.Contains(domain, "gitlab") {
		return gitlab.NewSource(gitlab.Config{BaseURL: domain})
	}
	return newGitHubSource(domain)
}

func newGitHubSource(domain string) (*github.Source, error) {
	config := github.Config{}
	if domain != "" && !strings.HasSuffix(domain, "://github.com") {
		config.EnterpriseBaseURL = domain
	}
	return github.NewSource(config)
}
