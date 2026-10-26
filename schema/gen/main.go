//go:build ignore

// Command gen writes the JSON Schema files in schema/ from the Go types.
// Run it from the repository root: go run ./schema/gen/main.go
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	coursespec "github.com/academyosio/course-spec"
)

const baseID = "https://academyos.io/schemas/course-spec/v1/"

func main() {
	docs := []struct {
		name  string
		title string
		v     any
	}{
		{"course", "course-spec course manifest", &coursespec.CourseManifest{}},
		{"section", "course-spec section manifest", &coursespec.SectionManifest{}},
		{"quiz", "course-spec quiz", &coursespec.Quiz{}},
		{"published", "course-spec published state", &coursespec.Published{}},
	}
	for _, d := range docs {
		r := &jsonschema.Reflector{ExpandedStruct: true}
		if err := r.AddGoComments("github.com/academyosio/course-spec", "./"); err != nil {
			log.Fatalf("read Go comments: %v", err)
		}
		s := r.Reflect(d.v)
		s.ID = jsonschema.ID(baseID + d.name + ".schema.json")
		s.Title = d.title
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			log.Fatalf("%s: %v", d.name, err)
		}
		path := filepath.Join("schema", d.name+".schema.json")
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			log.Fatalf("%s: %v", path, err)
		}
	}
}
