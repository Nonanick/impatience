package require

import (
	"github.com/nonanick/impatience/analyzer/javascript"
	"github.com/nonanick/impatience/transform"
)

// Register this file transformer to all .js files
// require will change "require()" from js scripts
// to import syntax, maybe i should just inject
// a request polyfill cause module.exports is another
// pain to transform
func Register() {
	transform.AddFileTransformer(".js", RequireTransform)
}

// RequireTransform function that analyzes
var RequireTransform transform.FileTransformer = func(
	path string,
	content []byte,
) []byte {

	matcher := javascript.RequireRegExp

	// Find all submatches using this matcher
	matches := matcher.FindAllSubmatch(content, -1)
	matchesIndex := matcher.FindAllIndex(content, -1)
	subNames := matcher.SubexpNames()

	if matches == nil {
		return content
	}

	// all transformation will be done in this
	newContent := []byte{}
	lastIndex := 0

	// Find all matches -- First one is the whole match!
	for ioSubMatch, subMatches := range matches {

		if len(subMatches) > 0 {
			name := subMatches[IndexOf("name", subNames)]
			path := subMatches[IndexOf("path", subNames)]

			startOfMatch := matchesIndex[ioSubMatch][0]
			endOfMatch := matchesIndex[ioSubMatch][1]

			newContent = append(newContent, content[lastIndex:startOfMatch]...)
			newContent = append(newContent, []byte("import "+string(name)+" from "+string(path)+";")...)
			lastIndex = endOfMatch

		}
	}

	newContent = append(newContent, content[lastIndex:]...)

	return newContent
}

// IndexOf index of a string in an array
func IndexOf(str string, arr []string) int {
	retInd := -1

	for i, s := range arr {
		if s == str {
			return i
		}
	}

	return retInd
}
