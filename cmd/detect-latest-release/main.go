package main

import (
	"flag"
	"fmt"
	"log"
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
	repo = strings.TrimPrefix(repo, "https://")
	repo = strings.TrimPrefix(repo, "github.com/")

	matched, err := regexp.MatchString("[^/]+/[^/]+", repo)
	if err != nil {
		panic(err)
	}
	if !matched {
		usage()
		os.Exit(1)
	}

	source := getSource(cvsType, repo)
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})

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
	fmt.Printf("Download URL: %s\n", latest.AssetURL)
	fmt.Printf("Release URL: %s\n", latest.URL)
	fmt.Printf("Release Notes:\n%s\n", latest.ReleaseNotes)
}

func getSource(cvsType, repo string) selfupdate.Source {
	source, _ := selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: "https://gitea.com/"})
	return source
}
