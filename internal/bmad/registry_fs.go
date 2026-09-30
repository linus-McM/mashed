package bmad

import "embed"

// registryFS holds the small CSV fixtures bundled with the binary for
// InputFromRegistry OptionsRef resolution. Keeping them in go:embed means
// production code paths and tests share the exact same data — no filesystem
// scavenger hunt, no test-only paths.
//
//go:embed testdata/*.csv
var registryFS embed.FS
