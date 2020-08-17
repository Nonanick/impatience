package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/nonanick/impatience/pathresolver"

	"github.com/nonanick/impatience/analyzer/css"
	"github.com/nonanick/impatience/analyzer/html"
	"github.com/nonanick/impatience/analyzer/javascript"
	"github.com/nonanick/impatience/crawler"
	"github.com/nonanick/impatience/impatienceserver"
)

func main() {

	// Get wd from running proccess
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		log.Fatalln("Working directory could not be reached!", wdErr)
	}

	// Build the absolute path from wd + given path
	absPath := filepath.Join(wd, "..", "public")

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
	knownFiles := crawler.AllFilenames(crawResult)
	knownFilesStatsArr := crawler.AllFiles(crawResult)
	var knownFilesStats = map[string]crawler.FileCrawlInfo{}

	for _, f := range knownFilesStatsArr {
		knownFilesStats[f.FilePath] = f
	}

	// Add configurations to HTTP2 server
	impatienceserver.Configure(
		&impatienceserver.ImpatienceConfig{
			Port:             443,
			Root:             absPath,
			KnownFiles:       knownFiles,
			KnownFilesStats:  &knownFilesStats,
			FileDependencies: &fileDependencies,
		},
	)

	// Add path resolvers - Absolute, Relative, With Index, With Extension
	pathresolver.AddResolver(pathresolver.Absolute)
	pathresolver.AddResolver(pathresolver.Relative)
	pathresolver.AddResolver(pathresolver.WithIndex)
	pathresolver.AddResolver(pathresolver.WithExtension)

	// Run Server
	impatienceserver.Launch()

}
