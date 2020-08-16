package main

import (
	"log"
	"os"
	"path"

	"github.com/nonanick/impatience/impatienceServer"

	"github.com/nonanick/impatience/analyzer/css"
	"github.com/nonanick/impatience/analyzer/html"
	"github.com/nonanick/impatience/analyzer/javascript"
	"github.com/nonanick/impatience/crawler"
)

func main() {

	// Get wd from running proccess
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		log.Fatalln("Working directory could not be reached!", wdErr)
	}

	// Build the absolute path from wd + given path
	absPath := path.Join(wd, "../public")

	// Crawl public directory
	crawResult := crawler.Crawl(absPath)

	// Add file analyzers
	javascript.Register()
	html.Register()
	css.Register()

	// Add file transformers
	javascript.UseNodeModules()

	// Generate dependencies map -- uses file analyzers, register them before calling
	fileDependencies := GenerateDependenciesMap(crawResult)

	// Add configurations to HTTP2 server
	impatienceServer.Configure(&impatienceServer.ImpatienceConfig{
		Port:             443,
		Root:             absPath,
		KnownFiles:       crawler.AllFilenames(crawResult),
		FileDependencies: fileDependencies,
	})

	// Run Server
	impatienceServer.Launch()

}
