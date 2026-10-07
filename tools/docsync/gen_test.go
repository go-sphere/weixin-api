package main

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/dave/jennifer/jen"
)

// TestNestedObjectsNamespacedByPath pins that same-named objects under
// different parents get distinct types instead of sharing the first shape.
func TestNestedObjectsNamespacedByPath(t *testing.T) {
	clear(generatedTypes)
	t.Cleanup(func() { clear(generatedTypes) })
	fields := []Field{
		{Name: "a", Type: "object", Fields: []Field{{Name: "item", Type: "object", Fields: []Field{{Name: "x", Type: "string"}}}}},
		{Name: "b", Type: "object", Fields: []Field{{Name: "item", Type: "object", Fields: []Field{{Name: "y", Type: "number"}}}}},
	}
	f := NewFile("p")
	f.Type().Id("Resp").StructFunc(func(g *Group) {
		used := make(map[string]int)
		for _, fl := range fields {
			writeField(f, g, fl, used, fieldTag("json", fl, fl.Required), "Resp")
		}
	})
	src := fmt.Sprintf("%#v", f)
	for _, want := range []string{"type RespAItemObject struct", "type RespBItemObject struct"} {
		if !strings.Contains(src, want) {
			t.Errorf("missing %q in:\n%s", want, src)
		}
	}
}

// TestPackageHeaderDocumentsPlatform pins a real package doc per platform.
func TestPackageHeaderDocumentsPlatform(t *testing.T) {
	h := packageHeader("official")
	if !strings.Contains(h, "// Package official is a generated client for the WeChat Official") || !strings.HasSuffix(h, "package official\n\n") {
		t.Fatalf("unexpected header:\n%s", h)
	}
}
