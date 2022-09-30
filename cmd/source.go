package cmd

import (
	"net/url"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
)

func GetSource(cvsType, repo string) (selfupdate.Source, string, error) {
	if !strings.HasPrefix(repo, "http") {
		source, err := getSourceFromName(cvsType)
		if err != nil {
			return nil, repo, err
		}
		return source, repo, nil
	}
	repoURL, err := url.Parse(repo)
	if err != nil {
		return nil, repo, err
	}
	slug := strings.TrimPrefix(repoURL.Path, "/")

	source, err := getSourceFromURL(repoURL)
	if err != nil {
		return nil, slug, err
	}
	return source, slug, nil
}

func getSourceFromName(name string) (selfupdate.Source, error) {
	switch name {
	case "gitea":
		return selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: "https://gitea.com/"})

	case "gitlab":
		return selfupdate.NewGitLabSource(selfupdate.GitLabConfig{})

	default:
		return selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	}
}

func getSourceFromURL(repoURL *url.URL) (selfupdate.Source, error) {
	if strings.Contains(repoURL.Hostname(), "gitea") {
		return selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: baseURL(repoURL)})
	}
	if strings.Contains(repoURL.Hostname(), "gitlab") {
		return selfupdate.NewGitLabSource(selfupdate.GitLabConfig{BaseURL: baseURL(repoURL)})
	}
	return selfupdate.NewGitHubSource(selfupdate.GitHubConfig{EnterpriseBaseURL: baseURL(repoURL)})
}

func baseURL(base *url.URL) string {
	return base.Scheme + "://" + base.Host
}
