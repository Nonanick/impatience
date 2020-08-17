package cache

import (
	"net/http"
	"strings"
)

// KnownFilesHashs All computed file hashes
var KnownFilesHashs = []string{}

const (
	// HashSalt Adds a static salt to hashed files
	HashSalt = "ImpatienceSvCookie"
	// CookieCachedFilesName name of the coockie holding the cache
	CookieCachedFilesName = "ImpatienceCacheControl"
	// CookieFileSeparator string that shall seoarate file names
	CookieFileSeparator = "_&_"
)

// CookieStrategy file cache strategy
var CookieStrategy = Strategy{
	// Extract strategy
	Extract: func(
		request *http.Request,
		knownFileHashes map[string]string,
	) []string {

		// Check if cookie exists, otherwise return empty string array
		cookie, err := request.Cookie(CookieCachedFilesName)
		if err != nil {
			return []string{}
		}

		fileHashs := strings.Split(cookie.Value, CookieFileSeparator)
		knownCaches := []string{}
		knownHashes := make(map[string]bool)

		// Create map of known hashs for easy access
		for _, knownHash := range knownFileHashes {
			knownHashes[knownHash] = true
		}

		// Check each cache hash and confirm if it exists in current known files (file change => invalidate cache!)
		for _, hash := range fileHashs {
			if knownHashes[hash] {
				knownCaches = append(knownCaches, hash)
			}
		}

		return knownCaches
	},
	// Insert strategy
	Insert: func(response http.ResponseWriter,
		previouslyCached []string,
		pushedFiles []string,
		knownFileHashs map[string]string,
	) {

	},
}
