package main

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"
)

// buildSwagger renders one generated package's operations as an OpenAPI 3.0
// document using the kin-openapi model, validates it against the OpenAPI
// specification, and returns JSON and YAML encodings. All identifier names
// derive from the operation ID, which itself derives from the endpoint path
// by fixed rules, so the contract is stable across runs.
func buildSwagger(ctx context.Context, title string, ops []*Operation) (jsonDoc, yamlDoc []byte, err error) {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:       title,
			Description: "Extracted from the official WeChat documentation by docsync. Do not edit; regenerate with make docs-sync.",
			Version:     "1.0.0",
		},
		Servers: openapi3.Servers{{URL: "https://api.weixin.qq.com"}},
		Paths:   openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
	}

	seen := make(map[string]bool)
	hasBinary := false
	for _, op := range ops {
		key := op.Method + " " + op.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		hasBinary = hasBinary || op.BinaryResponse

		item := doc.Paths.Find(op.Path)
		if item == nil {
			item = &openapi3.PathItem{}
		}
		pathOp := renderOperation(op)
		switch op.Method {
		case "GET":
			item.Get = pathOp
		case "POST":
			item.Post = pathOp
		case "PUT":
			item.Put = pathOp
		case "DELETE":
			item.Delete = pathOp
		default:
			item.Patch = pathOp
		}
		doc.Paths.Set(op.Path, item)

		doc.Components.Schemas[op.ID+"Request"] = &openapi3.SchemaRef{
			Value: objectSchema(op.Body),
		}
		if !op.BinaryResponse {
			doc.Components.Schemas[op.ID+"Response"] = &openapi3.SchemaRef{
				Value: objectSchema(op.Response),
			}
		}
		if op.Upload != nil {
			doc.Components.Schemas[op.ID+"Request"] = &openapi3.SchemaRef{
				Value: objectSchema(op.Body),
			}
		}
	}
	if hasBinary {
		doc.Components.Schemas["ErrorEnvelope"] = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"errcode": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}, Description: "错误码"}},
					"errmsg":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Description: "错误信息"}},
				},
			},
		}
	}

	// kin-openapi's Validate does not dereference $refs (that is the
	// Loader's job); link in-memory refs to their components targets so
	// validation passes. Marshaling still emits the plain $ref.
	resolveRefs(doc)

	if err := doc.Validate(ctx); err != nil {
		return nil, nil, err
	}

	jsonDoc, err = json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	jsonDoc = append(jsonDoc, '\n')
	yamlDoc, err = yaml.Marshal(doc)
	if err != nil {
		return nil, nil, err
	}
	return jsonDoc, yamlDoc, nil
}

// resolveRefs fills the Value of every local schema $ref used by the
// generated operations so kin-openapi's Validate can walk the document.
func resolveRefs(doc *openapi3.T) {
	resolve := func(ref *openapi3.SchemaRef) {
		if ref == nil || ref.Ref == "" || ref.Value != nil {
			return
		}
		name := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
		if target, ok := doc.Components.Schemas[name]; ok {
			ref.Value = target.Value
		}
	}
	for _, item := range doc.Paths.Map() {
		for _, o := range item.Operations() {
			if o.RequestBody != nil && o.RequestBody.Value != nil {
				for _, mt := range o.RequestBody.Value.Content {
					resolve(mt.Schema)
				}
			}
			for _, pr := range o.Parameters {
				if pr.Value != nil {
					resolve(pr.Value.Schema)
				}
			}
			for _, rr := range o.Responses.Map() {
				if rr.Value != nil {
					for _, mt := range rr.Value.Content {
						resolve(mt.Schema)
					}
				}
			}
		}
	}
}

// errCodeDoc is one documented error code, exposed on operations via the
// x-error-codes extension.
type errCodeDoc struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	Solution    string `json:"solution,omitempty"`
}

func renderOperation(op *Operation) *openapi3.Operation {
	o := &openapi3.Operation{
		OperationID: op.ID,
		Summary:     op.Summary,
		Description: "Doc: " + op.Page,
	}
	if len(op.Errors) > 0 {
		docs := make([]errCodeDoc, 0, len(op.Errors))
		for _, e := range op.Errors {
			docs = append(docs, errCodeDoc{Code: e.Code, Description: e.Desc, Solution: e.Solution})
		}
		o.Extensions = map[string]any{"x-error-codes": docs}
	}
	if op.BinaryResponse {
		// success returns raw bytes (image/media); failure returns the JSON
		// errcode envelope
		binaryDesc := op.Summary + ". 成功时直接返回二进制内容; 失败时返回 JSON 错误信封。"
		o.Responses = openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
				Description: &binaryDesc,
				Content: openapi3.Content{
					"application/octet-stream": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
							Type:   &openapi3.Types{"string"},
							Format: "binary",
						}},
					},
				},
			}}),
		)
		o.Responses.Set("default", &openapi3.ResponseRef{Value: &openapi3.Response{
			Description: new("请求失败时返回"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: schemaRef("ErrorEnvelope"),
				},
			},
		}})
		return o
	}
	o.Responses = openapi3.NewResponses(
		openapi3.WithStatus(200, &openapi3.ResponseRef{Value: &openapi3.Response{
			Description: &op.Summary,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: schemaRef(op.ID + "Response"),
				},
			},
		}}),
	)
	for _, f := range op.Query {
		o.Parameters = append(o.Parameters, &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name:        f.Name,
				In:          "query",
				Description: f.Desc,
				Required:    f.Required,
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:   &openapi3.Types{openAPIType(f.Type)},
						Format: openAPIFormat(f.Type),
					},
				},
			},
		})
	}
	sortParameters(o.Parameters)
	switch {
	case op.Upload != nil:
		// multipart/form-data: the file part plus any additional form fields
		props := openapi3.Schemas{
			op.Upload.Name: {Value: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Format:      "binary",
				Description: op.Upload.Desc,
			}},
		}
		for _, f := range op.Body {
			props[f.Name] = &openapi3.SchemaRef{Value: fieldSchema(f)}
		}
		o.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Description: "请求体 (multipart/form-data)",
				Required:    true,
				Content: openapi3.Content{
					"multipart/form-data": &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{
							Type:       &openapi3.Types{"object"},
							Properties: props,
						}},
					},
				},
			},
		}
	case len(op.Body) > 0:
		o.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Description: "请求体",
				Required:    true,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: schemaRef(op.ID + "Request"),
					},
				},
			},
		}
	}

	return o
}

func schemaRef(name string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
}

// objectSchema renders a field list as an object schema. Nested objects and
// arrays of objects collapse to free-form objects (the docs define their
// members in separate prose tables); the description keeps the original text.
func objectSchema(fields []Field) *openapi3.Schema {
	s := &openapi3.Schema{
		Type:       &openapi3.Types{"object"},
		Properties: make(openapi3.Schemas, len(fields)),
	}
	for _, f := range fields {
		s.Properties[f.Name] = &openapi3.SchemaRef{Value: fieldSchema(f)}
		if f.Required {
			s.Required = append(s.Required, f.Name)
		}
	}
	s.Required = slices.Sorted(slices.Values(s.Required))
	return s
}

// fieldSchema renders one field, inlining nested object members that the docs
// spell out in separate payload blocks.
func fieldSchema(f Field) *openapi3.Schema {
	if f.Type == "array" {
		items := &openapi3.Schema{Type: &openapi3.Types{"string"}}
		if len(f.Fields) > 0 {
			items = objectSchema(f.Fields)
		}
		return &openapi3.Schema{
			Type:        &openapi3.Types{"array"},
			Description: f.Desc,
			Items:       &openapi3.SchemaRef{Value: items},
		}
	}
	if len(f.Fields) > 0 {
		nested := objectSchema(f.Fields)
		nested.Description = f.Desc
		return nested
	}
	return &openapi3.Schema{
		Type:        &openapi3.Types{openAPIType(f.Type)},
		Format:      openAPIFormat(f.Type),
		Description: f.Desc,
	}
}

// openAPIType maps an extracted scalar type onto its OpenAPI primitive.
// Values outside the OpenAPI scalar set (a stray "formdata" in a response
// table, for example) degrade to string.
func openAPIType(t string) string {
	switch t {
	case "integer", "number", "boolean", "array", "object":
		return t
	default:
		return "string"
	}
}

// openAPIFormat adds the canonical format for numeric primitives.
func openAPIFormat(t string) string {
	switch t {
	case "integer":
		return "int64"
	case "number":
		return "double"
	default:
		return ""
	}
}

// parameterRank gives a parameter its position group: unreferenced entries
// first, then query parameters, then everything else. Keeping query parameters
// ahead of the rest matches how the documentation orders them, and within a
// group the name decides.
func parameterRank(p *openapi3.ParameterRef) (int, string) {
	if p.Value == nil {
		return 0, ""
	}
	if p.Value.In == "query" {
		return 1, p.Value.Name
	}
	return 2, p.Value.Name
}

// sortParameters orders parameters deterministically for the emitted contract.
func sortParameters(params openapi3.Parameters) {
	slices.SortStableFunc(params, func(a, b *openapi3.ParameterRef) int {
		aRank, aName := parameterRank(a)
		bRank, bName := parameterRank(b)
		if aRank != bRank {
			return cmp.Compare(aRank, bRank)
		}
		return cmp.Compare(aName, bName)
	})
}
