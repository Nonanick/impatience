package cache

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"

	"github.com/nonanick/impatience/crawler"
)

// Extract return all cached files passed in the HTTP request
// strategy should ignore now invalid cached files
func Extract(
	request *http.Request,
	allKnownFileHashs map[string]string,
) []string {
	return extractCacheStrategy(
		request,
		allKnownFileHashs,
	)
}

// Insert push all cached files (excluding now invalid ones) to the response
// using the defined strategy
func Insert(
	response http.ResponseWriter,
	previouslyCached []string,
	pushedFiles []string,
	knownFilesHashs map[string]string,
) {
	insertCacheStrategy(
		response,
		previouslyCached,
		pushedFiles,
		knownFilesHashs,
	)
}

// ChangeStrategy Change the functions for extracting and inserting cache
// information into the request/response
func ChangeStrategy(strategy Strategy) {
	extractCacheStrategy = strategy.Extract
	insertCacheStrategy = strategy.Insert
}

// ExtractStrategyFn function that will extract the cached files
// coming from the request
type ExtractStrategyFn func(
	request *http.Request,
	knownFiles map[string]string,
) []string

// InsertStrategyFn function that will output the cached information into
// a response
type InsertStrategyFn func(
	response http.ResponseWriter,
	previouslyCached []string,
	servedFiles []string,
	knownFileHashs map[string]string,
)

// Strategy define a cache strategy to be used by Impatience
type Strategy struct {
	Extract ExtractStrategyFn
	Insert  InsertStrategyFn
}

// CalculateHashs calculate hash using salt + file name + last modified time
func CalculateHashs(files []crawler.FileCrawlInfo) map[string]string {
	var allHashs = make(map[string]string)

	for _, file := range files {
		allHashs[file.FilePath] = CalculateHash(file)
	}

	return allHashs
}

// CalculateHash calculate the hash for a single file
func CalculateHash(file crawler.FileCrawlInfo) string {
	hash := md5.New()

	bytes := []byte(HashSalt + file.FilePath + file.LastModified)

	return base64.StdEncoding.EncodeToString(hash.Sum(bytes))
}

// Extract strategy currently being used
var extractCacheStrategy ExtractStrategyFn = CookieStrategy.Extract

// Insert strategy currently being used
var insertCacheStrategy InsertStrategyFn = CookieStrategy.Insert
