package main

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// Operation is the semantic definition of one WeChat server API, extracted
// from its documentation page. Every identifier is derived from the endpoint
// itself (path, page slug) by fixed rules — never from LLM output — so the
// generated API surface stays stable across runs.
type Operation struct {
	ID             string   // stable Go identifier, e.g. GetCgiBinToken
	Tree           string   // documentation tree: miniprogram | service | subscription
	Method         string   // HTTP verb
	Path           string   // request path, e.g. /cgi-bin/stable_token
	Summary        string   // Chinese title of the doc page
	Page           string   // canonical doc page URL
	Query          []Field  // query string parameters
	Body           []Field  // request body fields
	Response       []Field  // response payload fields
	Errors         []ErrRow // error code table
	BinaryResponse bool     // true when success returns raw bytes (image/media) and only failures return JSON
	Upload         *Field   // non-nil when the endpoint takes a multipart/form-data file part
	Auth           bool     // true when the endpoint takes an access_token (URL or parameter table)
	// CallerToken marks endpoints whose access_token parameter is the *user's*
	// OAuth token supplied by the caller, not the application access token the
	// runtime manages. The /sns/ web-authorization family works this way: the
	// caller exchanges a code for a user token and then presents it.
	CallerToken bool
	Signing     string // "" | "pay_sig" | "pay_sig_session": how the client must sign the request
}

// snsPathPrefix marks the web-authorization family, whose requests carry the
// user's OAuth token rather than the application access token.
const snsPathPrefix = "/sns/"

// Signing modes. XPay (虚拟支付) endpoints authenticate the request body itself
// with HMAC-SHA256 signatures the docs declare as query parameters, so the
// client — not the caller — must produce them.
const (
	// SignNone means the endpoint needs no request signature.
	SignNone = ""
	// SignPaySig means pay_sig = HMAC-SHA256(AppKey, path + "&" + rawBody) is
	// required (every /xpay/ endpoint).
	SignPaySig = "pay_sig"
	// SignPaySigSession additionally requires
	// signature = HMAC-SHA256(session_key, rawBody) for user-level calls.
	SignPaySigSession = "pay_sig_session"
)

// xpayPathPrefix marks the virtual-payment family, whose requests are signed
// with the merchant AppKey.
const xpayPathPrefix = "/xpay/"

// signingFor classifies an endpoint's request signing requirement. The rule is
// fixed: every /xpay/ endpoint is signed, and the user-level signature is
// required exactly when the docs list a "signature" query parameter beside
// pay_sig.
func signingFor(op *Operation) string {
	if !strings.HasPrefix(op.Path, xpayPathPrefix) {
		return SignNone
	}
	if slices.ContainsFunc(op.Query, func(q Field) bool { return q.Name == "signature" }) {
		return SignPaySigSession
	}
	return SignPaySig
}

// clientOwnedParams are query parameters the generated client supplies itself
// rather than exposing on the request struct: the access token comes from the
// token cache, and pay_sig on signed endpoints is computed from Config.AppKey.
// A struct field for them would let a zero value silently overwrite the
// credential the runtime just produced.
func clientOwnedParams(op *Operation, name string) bool {
	if name == "access_token" {
		// The /sns/ family expects the caller's OAuth token in this parameter,
		// so it stays on the request struct and no app token is injected.
		return !op.CallerToken
	}
	if name == "pay_sig" && op.Signing != SignNone {
		return true
	}
	return false
}

// sessionKeyField is the synthetic request field carrying the user session key
// used for user-level signatures. It is never serialized.
const sessionKeyField = "SessionKey"

// binaryPathPrefixes lists endpoints whose documented contract is a media
// download: success streams raw bytes (image/voice/video) and only failures
// carry the JSON errcode envelope. The doc prose alone is inconsistent for
// these ("获取临时素材" never says 二进制), so the media getters are pinned.
var binaryPathPrefixes = []string{
	"/cgi-bin/media/get",
	"/cgi-bin/media/getfeedbackmedia",
	"/cgi-bin/material/get_material",
}

// isBinaryPath reports whether the endpoint is a pinned media download.
func isBinaryPath(path string) bool {
	return slices.ContainsFunc(binaryPathPrefixes, func(p string) bool {
		return strings.HasPrefix(path, p)
	})
}

// treeOf maps a documentation page URL onto its tree, matching the seed
// prefixes in crawl.go.
func treeOf(pageURL string) string {
	path := mustPath(pageURL)
	for _, p := range []struct{ prefix, tree string }{
		{"/miniprogram/dev/server/API/", "miniprogram"},
		{"/doc/service/api/", "service"},
		{"/doc/subscription/api/", "subscription"},
	} {
		if strings.HasPrefix(path, p.prefix) {
			return p.tree
		}
	}
	return ""
}

// Field is one parameter or response field. When the docs define the field's
// members in a nested "Xxx Object Payload" block, Fields carries them;
// ItemsType records the element type for arrays.
type Field struct {
	Name      string
	Type      string // normalized: string | number | boolean | array | object
	ItemsType string // for arrays: the element type (string | number | boolean | object)
	Required  bool
	Desc      string
	Fields    []Field // nested members (object, or array of object)
}

// objectDef is one "Xxx Object Payload" block: the members of a nested object
// and whether the field is an array of that object.
type objectDef struct {
	isArray bool
	fields  []Field
}

// objectDefs maps a normalized dotted path ("phone_info.watermark") onto its
// nested definition.
type objectDefs map[string]objectDef

// objectPayloadPrefixes are the documented prefixes for nested definitions:
// Res.* belongs to the response, Body.* to the request body.
const (
	resPrefix  = "Res"
	bodyPrefix = "Body"
)

// parseObjectPayloadHeading recognizes headings like
// "Res.device_list(Array) Object Payload" and returns the normalized dotted
// path and whether the field is an array.
func parseObjectPayloadHeading(heading, prefix string) (path string, isArray bool, ok bool) {
	const suffix = "Object Payload"
	h := strings.TrimSpace(heading)
	if !strings.HasSuffix(h, suffix) {
		return "", false, false
	}
	head := strings.TrimSpace(strings.TrimSuffix(h, suffix))
	p, found := strings.CutPrefix(head, prefix+".")
	if !found {
		return "", false, false
	}
	p = strings.TrimSpace(p)
	if strings.HasSuffix(p, "(Array)") {
		isArray = true
		p = strings.TrimSpace(strings.TrimSuffix(p, "(Array)"))
	}
	if p == "" {
		return "", false, false
	}
	return p, isArray, true
}

// isObjectPayloadSection reports whether a heading introduces a nested object
// definition block ("Body.tag Object Payload"). Such blocks are siblings of
// the parameter tables, not additional parameter tables, and must be skipped
// when collecting a section's own fields.
func isObjectPayloadSection(heading string) bool {
	return strings.HasSuffix(strings.TrimSpace(heading), "Object Payload")
}

// collectObjectDefs gathers every nested definition block under sec for the
// given prefix.
func collectObjectDefs(sec *Section, prefix string) objectDefs {
	defs := objectDefs{}
	var walk func(s *Section)
	walk = func(s *Section) {
		if p, isArray, ok := parseObjectPayloadHeading(s.Heading, prefix); ok {
			if _, dup := defs[p]; !dup {
				defs[p] = objectDef{isArray: isArray, fields: fieldsFromTables(tablesInSection(s))}
			}
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(sec)
	return defs
}

// buildNested attaches the nested members to the top-level fields whose
// definitions were found, recursing into deeper paths.
func buildNested(fields []Field, defs objectDefs, prefix string) []Field {
	out := slices.Clone(fields)
	for i, f := range fields {
		nf := f
		path := f.Name
		if prefix != "" {
			path = prefix + "." + f.Name
		}
		def, ok := defs[path]
		if !ok {
			continue
		}
		switch f.Type {
		case "array":
			nf.ItemsType = "object"
			nf.Fields = buildNested(def.fields, defs, path)
		case "object":
			nf.Fields = buildNested(def.fields, defs, path)
		default:
			// the docs typed the parent loosely but it has a member block
			nf.Type = "object"
			if def.isArray {
				nf.Type = "array"
				nf.ItemsType = "object"
			}
			nf.Fields = buildNested(def.fields, defs, path)
		}
		out[i] = nf
	}
	return out
}

// topLevelFields returns the fields of sec's own payload table, preferring the
// explicit "返回体"/"请求体" child block over the section table itself.
func topLevelFields(sec *Section) []Field {
	for _, c := range sec.Children {
		t := sectionTitle(c.Heading)
		lower := strings.ToLower(t)
		if strings.Contains(t, "返回体") || strings.Contains(lower, "response payload") ||
			strings.Contains(t, "请求体") || strings.Contains(lower, "request payload") {
			if c.Table != nil {
				return fieldsFromTable(c.Table)
			}
		}
	}
	if sec.Table != nil {
		return fieldsFromTable(sec.Table)
	}
	for _, c := range sec.Children {
		if c.Table != nil {
			return fieldsFromTable(c.Table)
		}
	}
	return nil
}

// fieldsFromTable converts one table into fields, skipping unrecognized rows.
func fieldsFromTable(t *Table) []Field {
	if t == nil || len(t.Headers) == 0 {
		return nil
	}
	var fields []Field
	seen := make(map[string]bool)
	for _, row := range t.Rows {
		f, ok := fieldFromRow(t.Headers, row)
		if !ok || seen[f.Name] {
			continue
		}
		seen[f.Name] = true
		fields = append(fields, f)
	}
	return fields
}

// ErrRow is one row of the error code table.
type ErrRow struct {
	Code     string
	Desc     string
	Solution string
}

// decimalHint matches descriptions that mark a numeric field as having a
// fractional part; everything else typed "number" is an integer in practice.
var decimalHint = regexp.MustCompile(`小数|浮点|保留\s*\d+\s*位|小数点|比例|百分比|坐标`)

// normalizeType maps the free-form doc type column onto the scalar types used
// throughout, without a description (LLM fallback path).
func normalizeType(raw string) string {
	return normalizeFieldType(raw, "")
}

// normalizeFieldType maps the doc type column plus its description onto a
// scalar type. "number" is treated as an integer unless the description says
// otherwise (lat/long have decimals, stray ratios are fractional), which keeps
// errcode-style fields as int64 instead of float64.
func normalizeFieldType(raw, desc string) string {
	t := strings.ToLower(strings.TrimSpace(raw))
	t = strings.TrimSuffix(t, "s") // "strings", "numbers" appear occasionally
	switch t {
	case "string", "str", "text", "buffer", "bytes", "binary", "timestamp", "-":
		return "string"
	case "formdata":
		return "formdata"
	case "double", "float":
		return "number"
	case "number", "int", "integer", "int32", "int64", "uint32", "long":
		if decimalHint.MatchString(desc) {
			return "number"
		}
		return "integer"
	case "bool", "boolean":
		return "boolean"
	case "objarray":
		return "array"
	case "numarray":
		return "array"
	case "array", "list", "set":
		return "array"
	case "object", "obj", "map", "json":
		return "object"
	default:
		if strings.HasPrefix(t, "array.") || strings.HasSuffix(t, "[]") {
			return "array"
		}
		return "string"
	}
}

// headerRole classifies a table header column.
type headerRole int

const (
	roleIgnore headerRole = iota
	roleName
	roleType
	roleRequired
	roleDesc
)

func headerRoleFor(h string) headerRole {
	switch strings.TrimSpace(h) {
	case "参数名", "参数", "属性", "名称", "字段", "name":
		return roleName
	case "类型", "返回类型", "type":
		return roleType
	case "必填", "required":
		return roleRequired
	case "描述", "说明", "含义", "详细介绍", "description":
		return roleDesc
	default:
		return roleIgnore
	}
}

// fieldFromRow maps a table row onto a Field using the header roles.
func fieldFromRow(headers []string, row []string) (Field, bool) {
	if len(headers) != len(row) {
		return Field{}, false
	}
	var f Field
	var name, typ string
	for i, h := range headers {
		switch headerRoleFor(h) {
		case roleName:
			name = strings.TrimSpace(row[i])
		case roleType:
			typ = row[i]
		case roleRequired:
			f.Required = isRequiredValue(row[i])
		case roleDesc:
			f.Desc = cleanDesc(row[i])
		}
	}
	if name == "" || !isPlainIdentifier(name) {
		return Field{}, false
	}
	f.Name = name
	f.Type = normalizeFieldType(typ, f.Desc)
	if f.Type == "array" {
		f.ItemsType = arrayElementType(typ, f.Desc)
	}
	return f, true
}

// arrayElementType decides the element type of a documented array field. The
// docs distinguish "numarray" (numeric) from "objarray" (objects, whose members
// are defined in a nested block) and plain "array" (string elements).
func arrayElementType(raw, desc string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "numarray":
		return "integer"
	case "objarray":
		return "object"
	}
	if decimalHint.MatchString(desc) {
		return "number"
	}
	return "string"
}

// isRequiredValue interprets the required column of the doc tables.
func isRequiredValue(v string) bool {
	switch strings.TrimSpace(v) {
	case "是", "true", "True", "yes", "Yes", "required", "必填":
		return true
	default:
		return false
	}
}

// cleanDesc collapses whitespace in a description cell.
func cleanDesc(s string) string {
	return asciiText(strings.Join(strings.Fields(s), " "))
}

// typographicASCII maps Unicode punctuation that editors report as "ambiguous"
// (visually confusable with ASCII) onto its ASCII equivalent.
var typographicASCII = map[rune]rune{
	'\u00a0': ' ',  // no-break space
	'\u3000': ' ',  // ideographic space
	'\u2018': '\'', // left single quotation mark
	'\u2019': '\'', // right single quotation mark
	'\u201c': '"',  // left double quotation mark
	'\u201d': '"',  // right double quotation mark
	'\u2013': '-',  // en dash
	'\u2014': '-',  // em dash
	'\u2212': '-',  // minus sign
	'\u00b7': '.',  // middle dot
}

// asciiText rewrites the ASCII-confusable punctuation an editor flags as
// "ambiguous unicode characters" into plain ASCII.
//
// The mirrored documentation is consumed as a machine-readable contract and as
// Go doc comments, both read in editors that highlight characters which look
// like ASCII (the fullwidth comma , for instance). The upstream Chinese
// punctuation is typographically correct for prose but is not meaningful in
// this reference material, so the ASCII form is emitted instead. CJK
// ideographs and CJK-only punctuation (。、) have no ASCII counterpart and are
// left untouched, so the text stays readable Chinese.
func asciiText(s string) string {
	if s == "" {
		return s
	}
	// Fast path: nothing to rewrite.
	if !needsASCIIRewrite(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 0xff01 && r <= 0xff5e: // fullwidth ASCII variants
			b.WriteRune(r - 0xfee0)
		default:
			if repl, ok := typographicASCII[r]; ok {
				b.WriteRune(repl)
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// needsASCIIRewrite reports whether s contains any character asciiText changes.
func needsASCIIRewrite(s string) bool {
	for _, r := range s {
		if r >= 0xff01 && r <= 0xff5e {
			return true
		}
		if _, ok := typographicASCII[r]; ok {
			return true
		}
	}
	return false
}

// isPlainIdentifier rejects cells that are prose or examples rather than
// parameter names (e.g. rows describing nested objects with long sentences).
func isPlainIdentifier(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-', r == '.':
		default:
			return false
		}
	}
	return true
}

// errRowFromRow maps an error-code table row.
func errRowFromRow(headers, row []string) (ErrRow, bool) {
	if len(headers) != len(row) {
		return ErrRow{}, false
	}
	var e ErrRow
	for i, h := range headers {
		switch strings.TrimSpace(h) {
		case "错误码", "errcode", "返回码", "code":
			e.Code = asciiText(strings.TrimSpace(row[i]))
		case "错误描述", "描述", "说明", "errmsg":
			e.Desc = cleanDesc(row[i])
		case "解决方案", "排查方法", "解决方案排查工具":
			e.Solution = cleanDesc(row[i])
		}
	}
	return e, e.Code != ""
}

// numberingRe matches the leading section numbering ("1." / "2、" / "3 ").
var numberingRe = regexp.MustCompile(`^\d+[.、)]?\s*`)

// sectionTitle normalizes a heading: strips decorative "#" and the leading
// numbering ("2. 请求参数" -> "请求参数").
func sectionTitle(h string) string {
	t := strings.TrimLeft(strings.TrimSpace(h), "# \t")
	return strings.TrimSpace(numberingRe.ReplaceAllString(t, ""))
}

// matchTitle reports whether the heading equals one of the variants.
func matchTitle(h string, variants ...string) bool {
	t := sectionTitle(h)
	for _, v := range variants {
		if strings.HasPrefix(t, v) {
			return true
		}
	}
	return false
}

// tablesInSection collects the tables of a section (including sub-sections)
// in document order.
func tablesInSection(sec *Section) []*Table {
	var out []*Table
	var walk func(s *Section)
	walk = func(s *Section) {
		if s.Table != nil {
			out = append(out, s.Table)
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(sec)
	return out
}

// fieldsFromTables converts tables with recognizable parameter headers into
// fields; tables without matching headers are skipped.
func fieldsFromTables(tables []*Table) []Field {
	var fields []Field
	seen := make(map[string]bool)
	for _, t := range tables {
		if len(t.Headers) == 0 {
			continue
		}
		for _, row := range t.Rows {
			f, ok := fieldFromRow(t.Headers, row)
			if !ok || seen[f.Name] {
				continue
			}
			seen[f.Name] = true
			fields = append(fields, f)
		}
	}
	return fields
}

// errorsFromTables converts error-code tables into rows.
func errorsFromTables(tables []*Table) []ErrRow {
	var out []ErrRow
	for _, t := range tables {
		if len(t.Headers) == 0 {
			continue
		}
		for _, row := range t.Rows {
			if e, ok := errRowFromRow(t.Headers, row); ok {
				out = append(out, e)
			}
		}
	}
	return out
}

// isIndexPage reports whether a page carries the catalogue layout (a list of
// APIs) rather than a single API definition.
func isIndexPage(root *Section) bool {
	// catalogue pages have many outbound links and no "调用方式" section
	var hasCall bool
	var walk func(s *Section)
	walk = func(s *Section) {
		if matchTitle(s.Heading, "调用方式") {
			hasCall = true
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(root)
	return !hasCall
}

// extractOperations turns one parsed doc page into zero or more operations.
// Pages following the unified template (调用方式/请求参数/返回参数/错误码) yield
// exactly one; catalogue pages yield none.
func extractOperations(pageURL, title string, root *Section) []*Operation {
	if isIndexPage(root) {
		return nil
	}

	var callSec, paramSec, respSec, errSec *Section
	var walk func(s *Section)
	walk = func(s *Section) {
		switch {
		case matchTitle(s.Heading, "调用方式"):
			if callSec == nil {
				callSec = s
			}
		case matchTitle(s.Heading, "请求参数", "Query String Parameters", "Request Payload", "请求体", "查询参数"):
			if paramSec == nil {
				paramSec = s
			}
		case matchTitle(s.Heading, "返回参数", "返回体", "Response Payload"):
			if respSec == nil {
				respSec = s
			}
		case matchTitle(s.Heading, "错误码"):
			if errSec == nil {
				errSec = s
			}
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(root)

	if callSec == nil {
		return nil
	}

	method, path := endpointFromSection(callSec)
	if path == "" {
		return nil
	}

	op := &Operation{
		ID:      "", // assigned by the caller's name assigner
		Method:  method,
		Path:    path,
		Summary: asciiText(title),
		Page:    pageURL,
	}
	if paramSec != nil {
		bodyDefs := collectObjectDefs(paramSec, bodyPrefix)
		for _, sub := range paramSec.Children {
			if isObjectPayloadSection(sub.Heading) {
				continue // nested definition, already captured by collectObjectDefs
			}
			title := sectionTitle(sub.Heading)
			lower := strings.ToLower(title)
			switch {
			case strings.Contains(title, "查询") || strings.Contains(lower, "query"):
				op.Query = append(op.Query, fieldsFromTables(tablesInSection(sub))...)
			case strings.Contains(title, "请求体") || strings.Contains(lower, "payload") || strings.Contains(lower, "body"):
				op.Body = append(op.Body, buildNested(topLevelFields(sub), bodyDefs, "")...)
			}
		}
		if len(op.Query) == 0 && len(op.Body) == 0 {
			op.Body = buildNested(topLevelFields(paramSec), bodyDefs, "")
			if len(op.Body) == 0 {
				op.Body = fieldsFromTables(tablesInSection(paramSec))
			}
		}
	}
	if respSec != nil {
		respDefs := collectObjectDefs(respSec, resPrefix)
		if len(respDefs) > 0 {
			op.Response = buildNested(topLevelFields(respSec), respDefs, "")
		}
		if len(op.Response) == 0 {
			op.Response = fieldsFromTables(tablesInSection(respSec))
		}
	}
	if errSec != nil {
		op.Errors = errorsFromTables(tablesInSection(errSec))
	}
	op.Query = dedupeFields(op.Query)
	op.Body = dedupeFields(op.Body)
	op.Response = dedupeFields(op.Response)

	// A "formdata" body field is the documented multipart file part; it is
	// not a JSON member, so lift it out of the body into Upload.
	kept := op.Body[:0]
	for _, f := range op.Body {
		if f.Type == "formdata" && op.Upload == nil {
			upload := f
			op.Upload = &upload
			continue
		}
		kept = append(kept, f)
	}
	op.Body = kept

	// access_token is documented either as a query parameter or inline in the
	// 调用方式 URL; endpoints authenticating another way (appid+secret) have
	// neither, and must not be sent a token.
	op.Auth = callURLCarriesToken(callSec) ||
		slices.ContainsFunc(op.Query, func(q Field) bool { return q.Name == "access_token" })

	op.BinaryResponse = isBinaryResponseDoc(root) || isBinaryPath(op.Path)
	// The /sns/ family carries the user's OAuth token: the caller supplies it
	// and the runtime must not inject (or fetch) an application token.
	if strings.HasPrefix(op.Path, snsPathPrefix) {
		op.CallerToken = op.Auth
		op.Auth = false
	}
	op.Signing = signingFor(op)
	applyTypeOverrides(op)
	return []*Operation{op}
}

// binaryResponseMarker is the fixed sentence the docs use for endpoints that
// return raw bytes on success and a JSON error envelope on failure (mini
// program code images, media downloads).
const binaryResponseMarker = "二进制内容"

// isBinaryResponseDoc reports whether the page declares a raw-bytes success
// payload.
func isBinaryResponseDoc(root *Section) bool {
	var text strings.Builder
	var walk func(s *Section)
	walk = func(s *Section) {
		for _, t := range s.Texts {
			text.WriteString(t)
			text.WriteByte(' ')
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(root)
	return strings.Contains(text.String(), binaryResponseMarker)
}

// dedupeFields drops repeated parameter names, keeping first occurrence. The
// docs often repeat a name across nested explanation tables (e.g. the fields
// of an array element), and they all describe the same parameter.
func dedupeFields(fields []Field) []Field {
	seen := make(map[string]bool, len(fields))
	out := fields[:0]
	for _, f := range fields {
		if seen[f.Name] {
			continue
		}
		seen[f.Name] = true
		out = append(out, f)
	}
	return out
}

// tokenQueryRe matches access_token used as a query key. The docs render the
// call URL as "?access_token = ACCESS_TOKEN" (spaces around '='), and a path
// segment may itself be named access_token (/sns/oauth2/access_token), so the
// trailing '=' is what distinguishes a credential parameter.
var tokenQueryRe = regexp.MustCompile(`\baccess_token\s*=`)

// callURLCarriesToken reports whether the endpoint's own call URL embeds an
// access_token query parameter. Only the URL line is inspected: the
// surrounding prose frequents mentions of access_token for other reasons (the
// appid+secret flows return one in the response).
func callURLCarriesToken(callSec *Section) bool {
	found := false
	var walk func(s *Section)
	walk = func(s *Section) {
		if found {
			return
		}
		for _, t := range s.Texts {
			if !tokenQueryRe.MatchString(t) {
				continue
			}
			if _, _, ok := parseEndpointLine(t); ok {
				found = true
				return
			}
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(callSec)
	return found
}

// endpointFromSection finds the METHOD + URL declaration inside the 调用方式
// section, preferring the HTTPS 调用 sub-section.
func endpointFromSection(callSec *Section) (method, path string) {
	for _, t := range callSec.Texts {
		if m, u, ok := parseEndpointLine(t); ok {
			return m, u
		}
	}
	var walk func(s *Section)
	walk = func(s *Section) {
		if path != "" {
			return
		}
		for _, txt := range s.Texts {
			if m, u, ok := parseEndpointLine(txt); ok {
				method, path = m, u
				return
			}
		}
		for _, c := range s.Children {
			walk(c)
		}
	}
	walk(callSec)
	return method, path
}

// parseEndpointLine parses "GET https://api.weixin.qq.com/x/y" or
// "POST /x/y?..." into its method and path (query stripped).
func parseEndpointLine(text string) (method, path string, ok bool) {
	for _, m := range findEndpointMatches(text) {
		u := trailingPunct.ReplaceAllString(m[2], "")
		if before, _, found := strings.Cut(u, "?"); found {
			u = before
		}
		if u == "" || u == "/" {
			continue
		}
		u = strings.TrimPrefix(u, "https://api.weixin.qq.com")
		u = strings.TrimPrefix(u, "http://api.weixin.qq.com")
		return strings.ToUpper(m[1]), u, true
	}
	return "", "", false
}

// typeOverride is a reviewed correction to a field whose documented type is
// wrong. ElementType applies to arrays and defaults to string.
type typeOverride struct {
	Type        string
	ElementType string
}

// typeOverrides corrects field types the upstream documentation states
// incorrectly. The keys are "METHOD /path field"; the entries are fixed and
// reviewed (never model output) so the contract stays reproducible, and each
// cites the evidence that the documented type is wrong.
var typeOverrides = map[string]typeOverride{
	// The response table says "string" but the payload is an array of IPs; the
	// published examples and the reference clients use a list.
	"GET /cgi-bin/getcallbackip ip_list": {Type: "array"},
	// The response table says only "array" without an element type, but the API
	// returns numeric tag ids (the reference client used []int).
	"GET /cgi-bin/user/info tagid_list": {Type: "array", ElementType: "integer"},
	// Same field in the batch variant, where the table leaves the type as "-".
	"POST /cgi-bin/user/info/batchget tagid_list": {Type: "array", ElementType: "integer"},
}

// applyTypeOverrides rewrites documented types that upstream gets wrong. It
// recurses into nested members, because a correction may be needed inside an
// object payload (user_info_list[].tagid_list).
func applyTypeOverrides(op *Operation) {
	var fix func(fields []Field)
	fix = func(fields []Field) {
		for i := range fields {
			key := op.Method + " " + op.Path + " " + fields[i].Name
			if ov, ok := typeOverrides[key]; ok {
				if fields[i].Type != ov.Type {
					fields[i].Type = ov.Type
					fields[i].Fields = nil
				}
				elem := ov.ElementType
				if elem == "" {
					elem = "string"
				}
				fields[i].ItemsType = elem
			}
			if len(fields[i].Fields) > 0 {
				fix(fields[i].Fields)
			}
		}
	}
	fix(op.Query)
	fix(op.Body)
	fix(op.Response)
}

// assignOperationIDs fills in every operation's stable ID. Names derive from
// the endpoint path (the interface itself), falling back to the page slug
// only to disambiguate collisions. The rules are fixed, so the same set of
// endpoints always yields the same names.
func assignOperationIDs(ops []*Operation) {
	used := make(map[string]int, len(ops))
	for _, op := range ops {
		base := goNameForEndpoint(op.Method, op.Path)
		n := used[base]
		used[base]++
		if n > 0 {
			base = fmt.Sprintf("%s%d", base, n+1)
		}
		op.ID = base
	}
}

// goNameForEndpoint derives the deterministic Go identifier for an endpoint:
// verb + CamelCase of the path segments (split on separators and digit
// boundaries). "GET /cgi-bin/stable_token" -> "GetCgiBinStableToken".
func goNameForEndpoint(method, path string) string {
	words := splitWords(strings.Trim(path, "/"))
	name := methodVerb(method)
	for _, w := range words {
		name += upperFirst(strings.ToLower(w))
	}
	if name == "" {
		return "CallAPI"
	}
	return name
}

func methodVerb(method string) string {
	switch method {
	case "GET":
		return "Get"
	case "POST":
		return "Post"
	case "PUT":
		return "Put"
	case "DELETE":
		return "Delete"
	case "PATCH":
		return "Patch"
	default:
		return "Call"
	}
}

// splitWords splits an identifier on separators and letter/digit boundaries.
func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case r == '_' || r == '-' || r == '.' || r == ' ' || r == '/':
			flush()
		case r >= '0' && r <= '9':
			// letter -> digit boundary starts a new word
			if cur.Len() > 0 && !endsWithDigit(cur.String()) {
				flush()
			}
			cur.WriteRune(r)
		default:
			// digit -> letter boundary starts a new word
			if endsWithDigit(cur.String()) {
				flush()
			}
			cur.WriteRune(r)
		}
	}
	flush()
	return words
}

func endsWithDigit(s string) bool {
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last >= '0' && last <= '9'
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// parseContentNode is a convenience wrapper used by crawl.go.
func parseOperations(pageURL, title string, content *html.Node) []*Operation {
	ops := extractOperations(pageURL, title, parseStructured(content))
	tree := treeOf(pageURL)
	for _, op := range ops {
		op.Tree = tree
	}
	return ops
}

// compareOperations orders operations by path then method, the canonical
// document order of the generated packages.
func compareOperations(a, b *Operation) int {
	if a.Path != b.Path {
		return cmp.Compare(a.Path, b.Path)
	}
	return cmp.Compare(a.Method, b.Method)
}

// collectOperations gathers the operations of all crawled pages, dropping
// duplicates (the same endpoint documented twice in one tree) and ordering
// the result deterministically by path then method.
func collectOperations(pages map[string]*page) []*Operation {
	seen := make(map[string]bool)
	var ops []*Operation
	for _, url := range slices.Sorted(maps.Keys(pages)) {
		for _, op := range pages[url].Ops {
			key := op.Tree + " " + op.Method + " " + op.Path
			if seen[key] {
				continue
			}
			seen[key] = true
			ops = append(ops, op)
		}
	}
	slices.SortFunc(ops, compareOperations)
	return ops
}

// treeOps is one generated package together with the operations that belong
// to it.
type treeOps struct {
	pkg treePackage
	Ops []*Operation
}

// treePackage describes one generated client package.
type treePackage struct {
	Dir   string // directory under -gen-dir
	Name  string // Go package name
	Title string // swagger info title
}

// packageSpecs defines the generated packages: the mini program tree alone,
// plus the service and subscription trees merged into one official package
// (they document the same account type and share the cgi-bin endpoint
// system, so the same endpoint documented twice appears once).
var packageSpecs = []struct {
	pkg   treePackage
	trees []string
}{
	{
		pkg:   treePackage{Dir: "miniprogram", Name: "miniprogram", Title: "WeChat Mini Program Server APIs"},
		trees: []string{"miniprogram"},
	},
	{
		pkg:   treePackage{Dir: "official", Name: "official", Title: "WeChat Official/Service Account Server APIs"},
		trees: []string{"service", "subscription"},
	},
}

// packageOperations assembles the generated packages from the extracted
// operations. Each package's IDs are assigned independently, because name
// uniqueness only matters within one package.
func packageOperations(ops []*Operation) []treeOps {
	grouped := groupByTree(ops)
	pkgs := make([]treeOps, 0, len(packageSpecs))
	for _, spec := range packageSpecs {
		var merged []*Operation
		for _, tree := range spec.trees {
			merged = append(merged, grouped[tree]...)
		}
		slices.SortFunc(merged, compareOperations)
		merged = dedupeEndpoints(merged)
		pkgs = append(pkgs, treeOps{pkg: spec.pkg, Ops: merged})
	}
	for _, p := range pkgs {
		assignOperationIDs(p.Ops)
	}
	return pkgs
}

// groupByTree partitions operations into their trees, sorted
// deterministically.
func groupByTree(ops []*Operation) map[string][]*Operation {
	grouped := make(map[string][]*Operation)
	for _, op := range ops {
		if op.Tree == "" {
			continue
		}
		grouped[op.Tree] = append(grouped[op.Tree], op)
	}
	for tree := range grouped {
		slices.SortFunc(grouped[tree], compareOperations)
	}
	return grouped
}

// dedupeEndpoints drops repeated (method, path) pairs, keeping the first
// occurrence of the sorted input.
func dedupeEndpoints(ops []*Operation) []*Operation {
	seen := make(map[string]bool, len(ops))
	out := ops[:0]
	for _, op := range ops {
		key := op.Method + " " + op.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, op)
	}
	return out
}
