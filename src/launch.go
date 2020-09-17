package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/nonanick/impatience/analyzer/css"
	"github.com/nonanick/impatience/analyzer/html"
	"github.com/nonanick/impatience/analyzer/javascript"
	"github.com/nonanick/impatience/crawler"
	"github.com/nonanick/impatience/options"
	"github.com/nonanick/impatience/pathresolver"
	"github.com/nonanick/impatience/server"
	"github.com/nonanick/impatience/transform/nodemodules"
	"github.com/nonanick/impatience/transform/typescript"
	"github.com/nonanick/impatience/watcher"
)

// Launch will launch a new https server
func Launch(args []string) {

	// Get wd from running proccess
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		log.Fatalln("Working directory could not be reached!", wdErr)
	}

	// Declare options
	options.Use(options.Default)

	// Build the absolute path from wd + given path
	absPath := filepath.Join(wd, "..", "public")
	options.PublicRoot = absPath
	options.NodeModulesRoot = filepath.Join("node_modules")

	// Add file analyzers
	javascript.Register()
	html.Register()
	css.Register()

	// Add file transformers
	typescript.Register()
	nodemodules.Register()

	// Crawl public directory and
	// -- add all the known files
	// -- apply all file transformers
	// -- use all analyzers to determine dependencies
	crawler.Crawl(absPath)

	// Add configurations to HTTP2 server
	server.Configure(
		&server.ImpatienceConfig{
			Port: 443,
			Root: absPath,
		},
	)

	// Add path resolvers - Absolute, Relative, With Index, With Extension
	pathresolver.AddResolver(pathresolver.Absolute)
	pathresolver.AddResolver(pathresolver.Relative)
	pathresolver.AddResolver(pathresolver.WithIndex)
	pathresolver.AddResolver(pathresolver.WithExtension)

	// Start fs watcher
	go watcher.Watch()

	// Run Server
	server.Launch()
}
