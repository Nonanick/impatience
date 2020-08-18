package cache

import (
	"net/http"
	"strings"
	"time"
)

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
	) map[string]bool {

		// Check if cookie exists, otherwise return empty string array
		cookie, err := request.Cookie(CookieCachedFilesName)
		if err != nil {
			return map[string]bool{}
		}

		fileHashs := strings.Split(cookie.Value, CookieFileSeparator)
		knownCaches := map[string]bool{}
		knownHashes := make(map[string]bool)

		// Create map of known hashs for easy access
		for _, knownHash := range knownFileHashes {
			knownHashes[knownHash] = true
		}

		// Check each cache hash and confirm if it exists in current known files (file change => invalidate cache!)
		for _, hash := range fileHashs {
			if knownHashes[hash] {
				knownCaches[hash] = true
			}
		}

		return knownCaches
	},
	// Insert strategy
	Insert: func(
		response http.ResponseWriter,
		previouslyCached map[string]bool,
		pushedFiles []string,
		knownFileHashs map[string]string,
	) {
		allHashs := []string{}

		for hash := range previouslyCached {
			allHashs = append(allHashs, hash)
		}

		allHashs = append(allHashs, pushedFiles...)

		cookie := http.Cookie{
			HttpOnly: true,
			Name:     CookieCachedFilesName,
			Value:    strings.Join(allHashs, CookieFileSeparator),
			Path:     "/",
			Secure:   true,
			Expires:  time.Now().Add(2 * 24 * time.Hour),
		}

		http.SetCookie(response, &cookie)
	},
}
