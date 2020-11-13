package selfupdate

import (
	"os"
	"testing"
)

func TestGitHubTokenEnv(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("because $GITHUB_TOKEN is not set")
	}

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
	if _, err := NewGitHubSource(GitHubConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitHub source with API token config")
	}
}

func TestGitHubTokenIsNotSet(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		defer os.Setenv("GITHUB_TOKEN", token)
	}
	os.Setenv("GITHUB_TOKEN", "")

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
}

func TestGitHubEnterpriseClientInvalidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

func TestGitHubEnterpriseClientValidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitHub source with valid URL")
	}
}
