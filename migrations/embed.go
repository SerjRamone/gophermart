// Package migrations ...
package migrations

import "embed"

//go:embed *.sql
var SQLFiles embed.FS
