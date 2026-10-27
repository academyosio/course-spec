package schema_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/invopop/jsonschema"

	coursespec "github.com/academyosio/course-spec"
	"github.com/academyosio/course-spec/schema"
)

func TestNamesAndGet(t *testing.T) {
	if got := strings.Join(schema.Names(), ","); got != "course,published,quiz,section" {
		t.Fatalf("names = %s", got)
	}
	for _, n := range schema.Names() {
		data, err := schema.Get(n)
		if err != nil {
			t.Fatal(err)
		}
		var v struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(data, &v); err != nil {
			t.Fatalf("%s: %v", n, err)
		}
		if v.ID != "https://academyos.io/schemas/course-spec/v1/"+n+".schema.json" {
			t.Errorf("%s: $id = %s", n, v.ID)
		}
	}
	if _, err := schema.Get("nope"); err == nil {
		t.Error("expected an error for an unknown schema")
	}
}

// TestUpToDate fails when the Go types changed without regenerating the
// schemas: run go generate ./schema.
func TestUpToDate(t *testing.T) {
	types := map[string]any{
		"course":    &coursespec.CourseManifest{},
		"section":   &coursespec.SectionManifest{},
		"quiz":      &coursespec.Quiz{},
		"published": &coursespec.Published{},
	}
	embeddedFiles := map[string][]byte{}
	for name := range types {
		data, err := schema.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		embeddedFiles[name] = data
	}
	t.Chdir("..") // read Go comments the way the generator does, from the root
	for name, v := range types {
		r := &jsonschema.Reflector{ExpandedStruct: true}
		if err := r.AddGoComments("github.com/academyosio/course-spec", "./"); err != nil {
			t.Fatal(err)
		}
		s := r.Reflect(v)
		var got map[string]any
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		var want map[string]any
		if err := json.Unmarshal(embeddedFiles[name], &want); err != nil {
			t.Fatal(err)
		}
		delete(want, "$id")
		delete(want, "title")
		delete(got, "$id")
		delete(got, "title")
		a, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s.schema.json is out of date; run go generate ./schema", name)
		}
	}
}
