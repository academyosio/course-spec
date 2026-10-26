// Package schema holds the course-spec JSON Schema files, generated from the
// Go types by schema/gen.
package schema

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:generate sh -c "cd .. && go run ./schema/gen/main.go"

//go:embed *.schema.json
var files embed.FS

// Names returns the schema names: course, published, quiz, section.
func Names() []string {
	entries, err := files.ReadDir(".")
	if err != nil {
		panic(fmt.Sprintf("read embedded schemas: %v", err)) // embedded at build time
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".schema.json"))
	}
	sort.Strings(names)
	return names
}

// Get returns the named schema, such as "course".
func Get(name string) ([]byte, error) {
	data, err := files.ReadFile(name + ".schema.json")
	if err != nil {
		return nil, fmt.Errorf("no schema %q (schemas: %s)", name, strings.Join(Names(), ", "))
	}
	return data, nil
}
