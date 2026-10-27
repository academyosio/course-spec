//go:build tools

// This file keeps the schema generator's dependency in go.mod, since the
// generator itself is excluded from builds.
package main

import _ "github.com/invopop/jsonschema"
