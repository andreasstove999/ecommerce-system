package db

import _ "embed"

// Schema holds the bootstrap SQL for integration tests and local development.
//
//go:embed schema.sql
var Schema string
