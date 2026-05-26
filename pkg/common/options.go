package common

// Options controls which structs are analyzed and how findings are produced.
type Options struct {
	Patterns         []string // -type globs matched against named-type names (nil = all)
	KeepTags         bool     // preserve struct field tags in rendered text
	IncludeGenerated bool     // analyze structs in generated files (default: skip them)
	SkipCachePadded  bool     // skip structs containing a golang.org/x/sys/cpu.CacheLinePad field
	RespectNolint    bool     // suppress findings on types carrying a recognized //nolint directive
	NolintLinters    []string // named //nolint tokens that trigger suppression (bare //nolint always counts)
}
