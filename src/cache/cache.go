package cache

import (
	"crypto/md5"
	"encoding/base64"
	"net/http"
)

// Extract return all cached files passed in the HTTP request
// strategy should ignore now invalid cached files
func Extract(
	request *http.Request,
	allKnownFileHashs map[string]string,
) map[string]bool {
	return extractCacheStrategy(
		request,
		allKnownFileHashs,
	)
}

// Insert push all cached files (excluding now invalid ones) to the response
// using the defined strategy
func Insert(
	response http.ResponseWriter,
	previouslyCached map[string]bool,
	pushedFiles []string,
) {
	insertCacheStrategy(
		response,
		previouslyCached,
		pushedFiles,
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
) map[string]bool

// InsertStrategyFn function that will output the cached information into
// a response
type InsertStrategyFn func(
	response http.ResponseWriter,
	previouslyCached map[string]bool,
	servedFiles []string,
)

// Strategy define a cache strategy to be used by Impatience
type Strategy struct {
	Extract ExtractStrategyFn
	Insert  InsertStrategyFn
}

// CalculateHash calculate the hash for a single file
func CalculateHash(path string, lastModified string) string {
	hash := md5.New()
	bytes := []byte(HashSalt + path + lastModified)
	hash.Write(bytes)
	encodedHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return encodedHash
}

// Extract strategy currently being used
var extractCacheStrategy ExtractStrategyFn = CookieStrategy.Extract

// Insert strategy currently being used
var insertCacheStrategy InsertStrategyFn = CookieStrategy.Insert
