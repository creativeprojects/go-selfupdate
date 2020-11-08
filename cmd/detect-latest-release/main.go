package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/creativeprojects/go-selfupdate/selfupdate"
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage: detect-latest-release {repo}\n\n  {repo} must be URL to GitHub repository or in 'owner/name' format.\n\n")
	flag.PrintDefaults()
}

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Display debugging information")

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

	latest, found, err := selfupdate.DetectLatest(repo)
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
