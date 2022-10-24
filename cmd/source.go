package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
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

func GetSource(cvsType, domain string) (selfupdate.Source, error) {
	if cvsType != "auto" && cvsType != "" {
		source, err := getSourceFromName(cvsType, domain)
		if err != nil {
			return nil, err
		}
		return source, nil
	}

	source, err := getSourceFromURL(domain)
	if err != nil {
		return nil, err
	}
	return source, nil
}

func getSourceFromName(name, domain string) (selfupdate.Source, error) {
	switch name {
	case "gitea":
		return selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: domain})

	case "gitlab":
		return selfupdate.NewGitLabSource(selfupdate.GitLabConfig{BaseURL: domain})

	default:
		return newGitHubSource(domain)
	}
}

func getSourceFromURL(domain string) (selfupdate.Source, error) {
	if strings.Contains(domain, "gitea") {
		return selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: domain})
	}
	if strings.Contains(domain, "gitlab") {
		return selfupdate.NewGitLabSource(selfupdate.GitLabConfig{BaseURL: domain})
	}
	return newGitHubSource(domain)
}

func newGitHubSource(domain string) (*selfupdate.GitHubSource, error) {
	config := selfupdate.GitHubConfig{}
	if domain != "" && !strings.HasSuffix(domain, "://github.com") {
		config.EnterpriseBaseURL = domain
	}
	return selfupdate.NewGitHubSource(config)
}
