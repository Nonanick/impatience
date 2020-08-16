package nodeModules

// NodeModulesRoot Determine the node_modules root folder (whitout the node_modules directory)
var NodeModulesRoot = "/mnt/26b0ac63-8c9c-4283-a1ee-72a745d122af/auria-api/node_modules"

// NodeLibraryPreffix preffiz is used on transformed js/html files to indicate
// that a library dependency
var NodeLibraryPreffix = "/lib://"

var registeredNodeLibraries []string

// AddNodeLibrary adds a node library
func AddNodeLibrary(name string) {

}

// RegisterNodeTransformers add node transformers for known extensions
func RegisterNodeTransformers() {

}
