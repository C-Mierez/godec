// Package apidoc provides embedded API documentation assets.
package apidoc

import _ "embed"

// APIDocsHTML is the embedded HTML content for the API documentation page.
//
//go:embed scalar.html
var APIDocsHTML []byte
