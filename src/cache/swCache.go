package cache

import (
	"io/ioutil"
	"regexp"

	"github.com/nonanick/impatience/transform"
)

// ServiceWorkerPublicLink define the public link to server the impatience SW Cache handler
var ServiceWorkerPublicLink = "/impatience://sw.js"

var swScript = func() string {
	return `
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
		navigator.serviceWorker.register('` + ServiceWorkerPublicLink + `')
			.then(registration => {
      		console.log('Impatience ServiceWorker registration successful with scope: ', registration.scope);
			})
			.catch(err => {
				console.error('Faile to register impatience service worker!!', err);
			});
  });
}
`
}

// AddSWScriptTransformer HTML File tranaformer that adds a script tag instructing the installation of
// impatience ServiceWorker
var AddSWScriptTransformer transform.FileTransformer = func(data *[]byte) []byte {

	trueData := *data

	// Try to find body closing tag
	bodyRegExp, _ := regexp.Compile("(</body>")
	bodyClosing := bodyRegExp.FindIndex(trueData)

	if len(bodyClosing) > 0 {
		return concatenateSwScriptToBytes(trueData, bodyClosing[0])
	}
	// Try to find head closing tag
	headerRegExp, _ := regexp.Compile("(</head>")
	headClosing := headerRegExp.FindIndex(trueData)

	if len(headClosing) > 0 {
		return concatenateSwScriptToBytes(trueData, headClosing[0])
	}

	// Concatenate in end of string
	concatBytes := []byte(string(trueData) + "<script>" + swScript() + "</script>")
	return concatBytes

}

func concatenateSwScriptToBytes(bytes []byte, index int) []byte {
	newBytes := string(bytes[:index]) + "<script>" + swScript() + "</script>" + string(bytes[index:])
	bytesFromNewStr := []byte(newBytes)
	return bytesFromNewStr
}

// AddSWFileTransformer Adds the SW solution to the Impatience Server,
// applying a cache solution will override previous ones!
func AddSWFileTransformer() {
	transform.AddFileTransformer(".html", AddSWScriptTransformer)
}

// CreateSWTransformedFile load the impatienceJSCache as a KnownFile in the defined Public URL
func CreateSWTransformedFile() transform.File {

	swBytes, _ := ioutil.ReadFile("./sw/impatienceJSCache.js")

	return transform.File{
		Path:     ServiceWorkerPublicLink,
		MimeType: "text/javascript",
		Bytes:    &swBytes,
	}
}
