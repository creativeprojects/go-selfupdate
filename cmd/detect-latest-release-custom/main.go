package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage: detect-latest-release {repo}\n\n  {repo} must be in 'owner/name' format.\n\n")
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

	source, _ := selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: "https://gitea.com/"})
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: nil,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Arm:       0,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	latest, found, err := updater.DetectLatest(flag.Arg(0))
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
