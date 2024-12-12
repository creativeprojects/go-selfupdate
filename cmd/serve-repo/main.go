// Simple implementation of a HTTP server to be used by the updater with the http source
package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"time"
)

func main() {
	var root, listen, prefix, fixedSlug string
	flag.StringVar(&root, "repo", "", "Root path of the file server")
	flag.StringVar(&listen, "listen", "localhost:9947", "IP address and port used for the HTTP server")
	flag.StringVar(&prefix, "path-prefix", "/repo", "Prefix to the root path of the HTTP server")
	flag.StringVar(&fixedSlug, "fixed-slug", "creativeprojects/resticprofile", "Only answer on this particular slug. When NOT specified the files will be served with the slug in the path.")
	flag.Parse()

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(root))
	pathPrefix := path.Join("/", prefix, fixedSlug)
	log.Print("listening on http://" + listen + pathPrefix)
	mux.Handle(pathPrefix+"/", http.StripPrefix(pathPrefix, WithLogging(fs)))
	server := http.Server{
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: 15 * time.Second,
	}
	_ = server.ListenAndServe()
}
