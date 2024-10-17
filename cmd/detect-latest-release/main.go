package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/cmd"
)

const (
	usageBloc = `
Usage: detect-latest-release [flags] {repository}

  {repository} can be:
    - URL to a repository
    - "owner/repository_name" couple separated by a "/"
    - numeric ID for Gitlab only

`
)

func usage() {
	fmt.Fprint(os.Stderr, usageBloc, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	var help, verbose bool
	var cvsType, forceOS, forceArch, baseURL string
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&verbose, "v", false, "Display debugging information")
	flag.StringVar(&cvsType, "t", "auto", "Version control: \"github\", \"gitea\", \"gitlab\" or \"http\"")
	flag.StringVar(&forceOS, "o", "", "OS name to use (windows, darwin, linux, etc)")
	flag.StringVar(&forceArch, "a", "", "CPU architecture to use (amd64, arm64, etc)")
	flag.StringVar(&baseURL, "u", "", "Base URL for VCS on http or dedicated instances")

	flag.Usage = usage
	flag.Parse()

	if help || flag.NArg() != 1 {
		usage()
		return
	}

	if verbose {
		selfupdate.SetLogger(log.New(os.Stdout, "", 0))
	}

	repo := flag.Arg(0)

	domain, slug, err := cmd.SplitDomainSlug(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if domain == "" && baseURL != "" {
		domain = baseURL
	}

	if verbose {
		fmt.Printf("slug %q on domain %q\n", slug, domain)
	}

	source, err := cmd.GetSource(cvsType, domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := selfupdate.Config{
		Source: source,
	}
	if forceOS != "" {
		cfg.OS = forceOS
	}
	if forceArch != "" {
		cfg.Arch = forceArch
	}
	updater, err := selfupdate.NewUpdater(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug(slug))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !found {
		fmt.Println("No release found")
		return
	}
	fmt.Printf("Latest version: %s\n", latest.Version())
	fmt.Printf("Download URL:   %q\n", latest.AssetURL)
	fmt.Printf("Release URL:    %q\n", latest.URL)
	fmt.Printf("Release Notes:\n%s\n", latest.ReleaseNotes)
}
