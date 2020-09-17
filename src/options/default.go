package options

// Default default Impatience options
var Default = ImpatienceOptions{
	CacheCookieName:        "_ImpatienceCache",
	CacheFilenameSeparator: "_&_",
	ExternalAnalyzers:      map[string]string{},
	ExternalTransformers:   map[string]string{},
	UseNodeModules:         true,
	NodeModulesRoot:        "node_modules",
	SearchForNodeModulesIn: []string{".js", ".ts", ".jsx", ".tsx", ".vue"},
	TLSCertificateFile:     "./ssl/cert.pem",
	TLSKeyFile:             "./ssl/key.pem",
	UseHotReload:           false,
	WatchFiles:             true,
}
