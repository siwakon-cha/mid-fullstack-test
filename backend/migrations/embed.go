package migrations

import "embed"

// FS contains the versioned SQL migrations used at application startup.
//
//go:embed *.sql
var FS embed.FS
