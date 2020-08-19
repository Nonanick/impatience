// Package options hold configurations that are used by the Impatience Server
package options

// PublicRoot hold the absolute path pointing to the folder that shall be
// served by Impatience
var PublicRoot string

// ServerAddr server adress
var ServerAddr string

// ServerPort in which port the server shall run
var ServerPort uint16

// CacheCookieName name of the cookie used by the cookie cache strategy
var CacheCookieName string

// CacheFilenameSeparator string that will separate cached etags
var CacheFilenameSeparator string

// TLSCertificateFile path to the TLS Certificate file
var TLSCertificateFile string

// TLSKeyFile path to the TLS Key File
var TLSKeyFile string

// ExternalTransformers hold file transformers that live outside Go
// the map key is the extension (with the dot) that shall use the transformer
// the value is the command to be run, the transformer must output the file
// contents to the process output and the ExitCode must be 0!
var ExternalTransformers map[string]string

// ExternalAnalyzers hold file analyzers that live outside Go
// the map key is the extension (with the dot) that shall be analyzed
// the value is the command to be run, the analyzer must output each
// dependency in the output separated by a new line "\n" and the ExitCode
// must be 0!
var ExternalAnalyzers map[string]string

// UseNodeModules instructs Impatience to expose the required node
// libraries using the fake URL /__impatience/node/:library
var UseNodeModules bool

// SearchForNodeModulesIn extensions that shall be analyzed AND transformed
// looking for node libraries and replacing the library name with the fake
// URL /__impatience/node/:library
var SearchForNodeModulesIn []string

// NodeModulesRoot specifies the absolute path to the node_modules folder
// that shall be used
var NodeModulesRoot string

// UseHotReload tell the server to emit events to clients connected to socket
// at fake path /__impatience/listen/fileChanges
var UseHotReload bool

// WatchFiles instruct the server to watch for file changes
// unless you want to reload the server each time you update a line of code
// this should remain as "true"!
var WatchFiles bool

// ImpatienceOptions Options to be used by Impatience Server
type ImpatienceOptions struct {
	PublicRoot string

	CacheCookieName        string
	CacheFilenameSeparator string

	TLSCertificateFile string
	TLSKeyFile         string

	ExternalTransformers map[string]string
	ExternalAnalyzers    map[string]string

	UseNodeModules         bool
	SearchForNodeModulesIn []string
	NodeModulesRoot        string

	UseHotReload bool
	WatchFiles   bool
}

// Use a set of options, if the value corresponds to the zero value
// it shall be ignored
func Use(options ImpatienceOptions) {
	if options.PublicRoot != "" {
		PublicRoot = options.PublicRoot
	}

	if options.CacheCookieName != "" {
		CacheCookieName = options.CacheCookieName
	}

	if options.CacheFilenameSeparator != "" {
		CacheFilenameSeparator = options.CacheFilenameSeparator
	}

	if options.TLSCertificateFile != "" {
		TLSCertificateFile = options.TLSCertificateFile
	}

	if options.TLSKeyFile != "" {
		TLSKeyFile = options.TLSKeyFile
	}

	if options.ExternalTransformers != nil {
		ExternalTransformers = options.ExternalTransformers
	}

	if options.ExternalAnalyzers != nil {
		ExternalAnalyzers = options.ExternalAnalyzers
	}

	if options.SearchForNodeModulesIn != nil {
		SearchForNodeModulesIn = options.SearchForNodeModulesIn
	}

	if options.NodeModulesRoot != "" {
		NodeModulesRoot = options.NodeModulesRoot
	}

	UseNodeModules = options.UseNodeModules
	UseHotReload = options.UseHotReload
	WatchFiles = options.WatchFiles
}
