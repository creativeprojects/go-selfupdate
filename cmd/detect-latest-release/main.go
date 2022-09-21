package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage: detect-latest-release {repo}\n\n  {repo} must be URL to GitHub repository or in 'owner/name' format.\n\n")
	flag.PrintDefaults()
}

func main() {
	var verbose bool
	var cvsType string
	flag.BoolVar(&verbose, "v", false, "Display debugging information")
	flag.StringVar(&cvsType, "t", "auto", "Version control: \"github\", \"gitea\" or \"gitlab\"")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}

	if verbose {
		selfupdate.SetLogger(log.New(os.Stdout, "", 0))
	}

	repo := flag.Arg(0)

	source, repo, err := getSource(cvsType, repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	matched, err := regexp.MatchString("[^/]+/[^/]+", repo)
	if err != nil {
		panic(err)
	}
	if !matched {
		usage()
		os.Exit(1)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	latest, found, err := updater.DetectLatest(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !found {
		fmt.Println("No release was found")
		return
	}
	fmt.Printf("Latest version: %s\n", latest.Version())
	fmt.Printf("Download URL:   %q\n", latest.AssetURL)
	fmt.Printf("Release URL:    %q\n", latest.URL)
	fmt.Printf("Release Notes:\n%s\n", latest.ReleaseNotes)
}

func getSource(cvsType, repo string) (selfupdate.Source, string, error) {
	if !strings.HasPrefix(repo, "http") {
		switch cvsType {
		case "gitea":
			source, err := selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: "https://gitea.com/"})
			if err != nil {
				return nil, repo, err
			}
			return source, repo, nil

		case "gitlab":
			source, err := selfupdate.NewGitLabSource(selfupdate.GitLabConfig{})
			if err != nil {
				return nil, repo, err
			}
			return source, repo, nil

		default:
			source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
			if err != nil {
				return nil, repo, err
			}
			return source, repo, nil
		}
	}
	repoURL, err := url.Parse(repo)
	if err != nil {
		return nil, repo, err
	}
	slug := strings.TrimPrefix(repoURL.Path, "/")

	if strings.Contains(repoURL.Hostname(), "gitea") {
		source, err := selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: baseURL(repoURL)})
		if err != nil {
			return nil, slug, err
		}
		return source, slug, nil
	}
	if strings.Contains(repoURL.Hostname(), "gitlab") {
		source, err := selfupdate.NewGitLabSource(selfupdate.GitLabConfig{BaseURL: baseURL(repoURL)})
		if err != nil {
			return nil, slug, err
		}
		return source, slug, nil
	}
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{EnterpriseBaseURL: baseURL(repoURL)})
	if err != nil {
		return nil, slug, err
	}
	return source, slug, nil
}

func baseURL(base *url.URL) string {
	return base.Scheme + "://" + base.Host
}
