package migrations

import "embed"

//go:embed *.sql
var Sql embed.FS
