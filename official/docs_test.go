package official

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestAllAPIEndpointsHaveReferenceURLs verifies every exported method on
// *OfficialAccount carries a "Reference:" URL pointing at the WeChat docs.
func TestAllAPIEndpointsHaveReferenceURLs(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var missing []string
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if !isOfficialAccountReceiver(fn.Recv) {
				continue
			}
			if fn.Name.Name == "AppID" {
				continue
			}
			if !hasReferenceURL(fn.Doc) {
				missing = append(missing, fset.Position(fn.Pos()).String()+" "+fn.Name.Name)
			}
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d exported API methods lack a Reference: doc URL:\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

// TestAllExportedTypesHaveDocComments verifies every exported type has a doc
// comment.
func TestAllExportedTypesHaveDocComments(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var missing []string
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				if ts.Doc == nil && gd.Doc == nil {
					missing = append(missing, fset.Position(ts.Pos()).String()+" "+ts.Name.Name)
				}
			}
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d exported types lack a doc comment:\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

func isOfficialAccountReceiver(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) != 1 {
		return false
	}
	star, ok := recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "OfficialAccount"
}

func hasReferenceURL(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	text := doc.Text()
	return strings.Contains(text, "Reference:") && strings.Contains(text, "https://")
}
