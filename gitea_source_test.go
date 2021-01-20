package selfupdate

import (
	"os"
	"testing"
)

func TestGiteaTokenEnv(t *testing.T) {
	token := os.Getenv("GITEA_TOKEN")
	if token == "" {
		t.Skip("because $GITEA_TOKEN is not set")
	}

	if _, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
	if _, err := NewGitHubSource(GitHubConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitHub source with API token config")
	}
}

func TestGiteaTokenIsNotSet(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		defer os.Setenv("GITHUB_TOKEN", token)
	}
	os.Setenv("GITHUB_TOKEN", "")

	if _, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
}
