package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// zhSampleDoc mirrors the unified template used by all three doc trees.
const zhSampleDoc = `<div class="content custom" data-v-6a143e3c>
<h1>获取接口调用凭据</h1>
<p>接口应在服务器端调用,不可在前端直接调用。</p>
<h2># 1. 调用方式</h2>
<h3># HTTPS 调用</h3>
<p>GET https://api.weixin.qq.com/cgi-bin/token?appid=APPID&amp;secret=SECRET&amp;grant_type=client_credential</p>
<h2># 2. 请求参数</h2>
<h3># 查询参数 Query String Parameters</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody>
<tr><td>access_token</td><td>string</td><td>是</td><td>接口调用凭证,有效期 7200 秒</td></tr>
<tr><td>grant_type</td><td>string</td><td>是</td><td>填写 client_credential</td></tr>
</tbody></table>
<h3># 请求体 Request Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>extra</td><td>string</td><td>附加信息</td></tr></tbody></table>
<h2># 3. 返回参数</h2>
<h3># 返回体 Response Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>access_token</td><td>string</td><td>获取到的凭证</td></tr></tbody></table>
<h2># 6. 错误码</h2>
<table><thead><tr><th>错误码</th><th>错误描述</th><th>解决方案</th></tr></thead>
<tbody><tr><td>40001</td><td>invalid credential</td><td>检查 appid</td></tr></tbody></table>
</div>`

const zhIndexDoc = `<div class="content custom">
<h1>小程序登录</h1>
<table><thead><tr><th>接口名称</th><th>请求路径</th><th>描述</th></tr></thead>
<tbody><tr><td>小程序登录凭证校验</td><td>/sns/jscode2session</td><td>登录凭证校验</td></tr></tbody></table>
</div>`

func TestExtractOperations(t *testing.T) {
	_, _, _, content, err := parsePage([]byte(wrapDoc(zhSampleDoc)), "https://developers.weixin.qq.com/doc/service/api/base/api_getaccesstoken.html")
	if err != nil {
		t.Fatalf("parsePage: %v", err)
	}
	ops := parseOperations("https://developers.weixin.qq.com/doc/service/api/base/api_getaccesstoken.html", "获取接口调用凭据", content)
	if len(ops) != 1 {
		t.Fatalf("ops = %d, want 1", len(ops))
	}
	op := ops[0]
	if op.Method != "GET" || op.Path != "/cgi-bin/token" {
		t.Errorf("endpoint = %s %s", op.Method, op.Path)
	}
	if len(op.Query) != 2 || op.Query[0].Name != "access_token" || !op.Query[0].Required {
		t.Errorf("query = %+v", op.Query)
	}
	if len(op.Body) != 1 || op.Body[0].Name != "extra" {
		t.Errorf("body = %+v", op.Body)
	}
	if len(op.Response) != 1 || op.Response[0].Type != "string" {
		t.Errorf("response = %+v", op.Response)
	}
	if len(op.Errors) != 1 || op.Errors[0].Code != "40001" {
		t.Errorf("errors = %+v", op.Errors)
	}

	// catalogue pages yield no operations
	_, _, _, content2, err := parsePage([]byte(wrapDoc(zhIndexDoc)), "https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/index.html")
	if err != nil {
		t.Fatalf("parsePage index: %v", err)
	}
	if ops := parseOperations("https://developers.weixin.qq.com/miniprogram/dev/server/API/user-login/index.html", "小程序登录", content2); len(ops) != 0 {
		t.Errorf("index page yielded %d ops, want 0", len(ops))
	}
}

func TestGoNameForEndpoint(t *testing.T) {
	tests := map[[2]string]string{
		{"GET", "/cgi-bin/stable_token"}:             "GetCgiBinStableToken",
		{"GET", "/sns/jscode2session"}:               "GetSnsJscode2Session",
		{"POST", "/wxa/sec/order/get_order"}:         "PostWxaSecOrderGetOrder",
		{"POST", "/cgi-bin/tags/update"}:             "PostCgiBinTagsUpdate",
		{"GET", "/cgi-bin/get_api_domain_ip"}:        "GetCgiBinGetApiDomainIp",
		{"POST", "/wxa/business/getuserphonenumber"}: "PostWxaBusinessGetuserphonenumber",
	}
	for ep, want := range tests {
		if got := goNameForEndpoint(ep[0], ep[1]); got != want {
			t.Errorf("goNameForEndpoint(%s %s) = %q, want %q", ep[0], ep[1], got, want)
		}
	}
}

func TestNormalizeType(t *testing.T) {
	tests := map[string]string{
		"string": "string", "String": "string", "number": "integer",
		"int32": "integer", "double": "number", "objarray": "array",
		"array.<string>": "array", "object": "object", "boolean": "boolean",
		"formdata": "formdata", "-": "string", "whatever": "string",
	}
	for in, want := range tests {
		if got := normalizeType(in); got != want {
			t.Errorf("normalizeType(%q) = %q, want %q", in, got, want)
		}
	}
	// descriptions that imply a fractional value keep the float type
	for desc, want := range map[string]string{
		"经度（小数点后6位）":        "number",
		"人均停留时长 (浮点型，单位：秒)": "number",
		"错误码": "integer",
		"":    "integer",
	} {
		if got := normalizeFieldType("number", desc); got != want {
			t.Errorf("normalizeFieldType(number, %q) = %q, want %q", desc, got, want)
		}
	}
}

func TestSwaggerOutputIsValidJSON(t *testing.T) {
	_, _, _, content, err := parsePage([]byte(wrapDoc(zhSampleDoc)), "https://developers.weixin.qq.com/doc/service/api/base/api_getaccesstoken.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://developers.weixin.qq.com/doc/service/api/base/api_getaccesstoken.html", "获取接口调用凭据", content)
	assignOperationIDs(ops)
	data, _, err := buildSwagger(t.Context(), "WeChat Server APIs", ops)
	if err != nil {
		t.Fatal(err)
	}
	if !utf8.Valid(data) {
		t.Fatal("swagger output is not valid UTF-8")
	}
	if !strings.Contains(string(data), `"openapi": "3.0.3"`) || !strings.Contains(string(data), `"/cgi-bin/token"`) {
		t.Errorf("swagger missing expected content:\n%s", data)
	}
}

func wrapDoc(inner string) string {
	return `<!DOCTYPE html><html><body><div id="docContent"><div class="content custom">` + inner + `</div></div></body></html>`
}

// Endpoints whose notes say success returns raw bytes must be flagged so the
// generated method returns ([]byte, error) and the contract declares a
// binary 200 with a default error envelope.
func TestBinaryResponseEndpoints(t *testing.T) {
	doc := `<!DOCTYPE html><html><body><div id="docContent"><div class="content custom">
<h1>获取小程序码</h1>
<h2>1. 调用方式</h2>
<h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=ACCESS_TOKEN</p>
<h2>2. 请求参数</h2>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody><tr><td>scene</td><td>string</td><td>是</td><td>场景值</td></tr></tbody></table>
<h2>4. 注意事项</h2>
<p>如果调用成功,会直接返回图片二进制内容,如果请求失败,会返回 JSON 格式的数据。</p>
</div></div></body></html>`
	_, _, _, content, err := parsePage([]byte(doc), "https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getwxacodeunlimit.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://developers.weixin.qq.com/miniprogram/dev/server/API/qrcode-link/qr-code/api_getwxacodeunlimit.html", "获取小程序码", content)
	if len(ops) != 1 || !ops[0].BinaryResponse {
		t.Fatalf("ops = %+v, want BinaryResponse", ops)
	}
	assignOperationIDs(ops)
	jsonDoc, _, err := buildSwagger(t.Context(), "test", ops)
	if err != nil {
		t.Fatal(err)
	}
	out := string(jsonDoc)
	for _, want := range []string{`"format": "binary"`, `"ErrorEnvelope"`, `"default"`} {
		if !strings.Contains(out, want) {
			t.Errorf("swagger missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, `GetWxaGetwxacodeunlimitResponse`) {
		t.Error("binary endpoint must not declare a JSON response struct")
	}
}

// Nested response members defined in separate "Res.x Object Payload" blocks
// must be attached to their parent field, including array-of-object and
// multi-level nesting.
func TestNestedObjectPayload(t *testing.T) {
	doc := `<!DOCTYPE html><html><body><div id="docContent"><div class="content custom">
<h1>获取手机号</h1>
<h2>1. 调用方式</h2>
<h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/wxa/business/getuserphonenumber</p>
<h2>3. 返回参数</h2>
<h3>返回体 Response Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>phone_info</td><td>object</td><td>用户手机号信息</td></tr></tbody></table>
<h3>Res.phone_info Object Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody>
<tr><td>phoneNumber</td><td>string</td><td>手机号</td></tr>
<tr><td>watermark</td><td>object</td><td>水印</td></tr>
</tbody></table>
<h3>Res.phone_info.watermark Object Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>appid</td><td>string</td><td>小程序appid</td></tr></tbody></table>
<h3>Res.list(Array) Object Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>id</td><td>number</td><td>编号</td></tr></tbody></table>
</div></div></body></html>`
	_, _, _, content, err := parsePage([]byte(doc), "https://developers.weixin.qq.com/miniprogram/dev/server/API/user-info/api_getphonenumber.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://x/user-info/api_getphonenumber.html", "获取手机号", content)
	if len(ops) != 1 || len(ops[0].Response) != 1 {
		t.Fatalf("ops = %+v", ops)
	}
	parent := ops[0].Response[0]
	if parent.Name != "phone_info" || len(parent.Fields) != 2 {
		t.Fatalf("parent = %+v", parent)
	}
	if parent.Fields[1].Name != "watermark" || len(parent.Fields[1].Fields) != 1 {
		t.Fatalf("watermark = %+v", parent.Fields[1])
	}
	if parent.Fields[1].Fields[0].Name != "appid" {
		t.Errorf("deep nesting lost: %+v", parent.Fields[1].Fields)
	}
}

// Endpoints that authenticate with appid+secret (or none) must not be flagged
// as needing an access_token: their call URL has no access_token query key.
func TestAuthDetection(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want bool
	}{
		{
			name: "token query key",
			doc: `<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/cgi-bin/callback/check?access_token = ACCESS_TOKEN</p>`,
			want: true,
		},
		{
			name: "appid secret flow",
			doc: `<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>GET https://api.weixin.qq.com/sns/jscode2session?appid = APPID &amp; secret = SECRET &amp; js_code = CODE</p>`,
			want: false,
		},
		{
			name: "path named access_token is not a credential",
			doc: `<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>GET https://api.weixin.qq.com/sns/oauth2/access_token?appid = APPID &amp; secret = SECRET &amp; code = CODE</p>`,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := wrapDoc("<h1>x</h1>" + tc.doc + `<h2>2. 请求参数</h2>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody><tr><td>p</td><td>string</td><td>是</td><td>d</td></tr></tbody></table>`)
			_, _, _, content, err := parsePage([]byte(doc), "https://developers.weixin.qq.com/doc/service/api/base/x.html")
			if err != nil {
				t.Fatal(err)
			}
			var tokens []*Operation
			for _, op := range parseOperations("https://x/api_base_x.html", "x", content) {
				tokens = append(tokens, op)
			}
			if len(tokens) != 1 {
				t.Fatalf("ops = %d", len(tokens))
			}
			if tokens[0].Auth != tc.want {
				t.Errorf("Auth = %v, want %v", tokens[0].Auth, tc.want)
			}
		})
	}
}

// A table whose header row carries extra columns (示例/枚举) must still be
// recognized: the header row is promoted by vocabulary, not by an exact
// column count, otherwise every row of the table is silently dropped.
func TestExtraHeaderColumns(t *testing.T) {
	doc := `<h1>网络通信检测</h1>
<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/cgi-bin/callback/check?access_token = ACCESS_TOKEN</p>
<h2>2. 请求参数</h2><h3>请求体 Request Payload</h3>
<table><tbody>
<tr><td>参数名</td><td>类型</td><td>必填</td><td>示例</td><td>说明</td><td>枚举</td></tr>
<tr><td>action</td><td>string</td><td>是</td><td>all</td><td>检测动作</td><td>-</td></tr>
<tr><td>check_operator</td><td>string</td><td>是</td><td>DEFAULT</td><td>检测运营商</td><td>CHINANET UNICOM</td></tr>
</tbody></table>`
	_, _, _, content, err := parsePage([]byte(wrapDoc(doc)), "https://developers.weixin.qq.com/miniprogram/dev/server/API/x.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://x/api_x.html", "网络通信检测", content)
	if len(ops) != 1 {
		t.Fatalf("ops = %d", len(ops))
	}
	got := ops[0].Body
	if len(got) != 2 || got[0].Name != "action" || !got[0].Required {
		t.Fatalf("body = %+v", got)
	}
	if got[1].Name != "check_operator" || got[1].Desc != "检测运营商" {
		t.Errorf("body[1] = %+v", got[1])
	}
}

// The override table corrects types upstream states wrongly: ip_list is
// documented as string but returned as an array, and user/info's tagid_list is
// documented as a bare array but returns numeric ids.
func TestTypeOverride(t *testing.T) {
	ip := typeOverrides["GET /cgi-bin/getcallbackip ip_list"]
	if ip.Type != "array" {
		t.Errorf("ip_list override missing: %+v", ip)
	}
	tag := typeOverrides["GET /cgi-bin/user/info tagid_list"]
	if tag.Type != "array" || tag.ElementType != "integer" {
		t.Errorf("tagid_list override wrong: %+v", tag)
	}
}

// numarray maps to a numeric slice, not []string.
func TestArrayElementTypes(t *testing.T) {
	cases := map[string]string{
		"numarray":  "integer",
		"objarray":  "object",
		"array":     "string",
		"numarray ": "integer",
		"":          "string",
	}
	for in, want := range cases {
		if got := arrayElementType(in, ""); got != want {
			t.Errorf("arrayElementType(%q) = %q, want %q", in, got, want)
		}
	}
}

// A nested "Body.x Object Payload" block sits beside the request-body table as
// a sibling heading; it must not be mistaken for another request-body table,
// which would duplicate its members at the top level.
func TestNestedBodyBlockNotTreatedAsBodyTable(t *testing.T) {
	doc := `<h1>创建标签</h1>
<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/cgi-bin/tags/create?access_token = ACCESS_TOKEN</p>
<h2>2. 请求参数</h2>
<h3>请求体 Request Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody><tr><td>tag</td><td>object</td><td>是</td><td>标签信息</td></tr></tbody></table>
<h3>Body.tag Object Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody><tr><td>name</td><td>string</td><td>是</td><td>标签名</td></tr></tbody></table>`
	_, _, _, content, err := parsePage([]byte(wrapDoc(doc)), "https://developers.weixin.qq.com/doc/service/api/x.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://x/api_x.html", "创建标签", content)
	if len(ops) != 1 {
		t.Fatalf("ops = %d", len(ops))
	}
	body := ops[0].Body
	if len(body) != 1 {
		t.Fatalf("body should hold only the top-level field, got %+v", body)
	}
	if body[0].Name != "tag" || len(body[0].Fields) != 1 || body[0].Fields[0].Name != "name" {
		t.Errorf("nested body wrong: %+v (fields %+v)", body[0], body[0].Fields)
	}
}

// Overrides must reach nested members too: batchget's tagid_list lives inside
// user_info_list[].
func TestTypeOverrideAppliesToNestedFields(t *testing.T) {
	doc := `<h1>批量获取用户信息</h1>
<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>POST https://api.weixin.qq.com/cgi-bin/user/info/batchget?access_token = ACCESS_TOKEN</p>
<h2>3. 返回参数</h2><h3>返回体 Response Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>user_info_list</td><td>objarray</td><td>用户信息列表</td></tr></tbody></table>
<h3>Res.user_info_list(Array) Object Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>说明</th></tr></thead>
<tbody><tr><td>tagid_list</td><td>-</td><td>用户被打上的标签ID列表</td></tr></tbody></table>`
	_, _, _, content, err := parsePage([]byte(wrapDoc(doc)), "https://developers.weixin.qq.com/doc/service/api/x.html")
	if err != nil {
		t.Fatal(err)
	}
	ops := parseOperations("https://x/api_x.html", "批量获取用户信息", content)
	if len(ops) != 1 || len(ops[0].Response) != 1 {
		t.Fatalf("ops = %+v", ops)
	}
	nested := ops[0].Response[0].Fields
	if len(nested) != 1 || nested[0].Name != "tagid_list" {
		t.Fatalf("nested = %+v", nested)
	}
	if nested[0].Type != "array" || nested[0].ItemsType != "integer" {
		t.Errorf("nested override not applied: %+v", nested[0])
	}
}

// The generated artifacts must not carry ASCII-confusable punctuation, which
// editors report as "ambiguous unicode characters". Chinese text is preserved.
func TestASCIITextNormalization(t *testing.T) {
	in := "接口调用凭证，可使用 access_token （见文档）：请勿外传；示例“小程序”±"
	got := asciiText(in)
	for _, bad := range []rune{'，', '（', '）', '：', '；', '“', '”'} {
		if strings.ContainsRune(got, bad) {
			t.Errorf("asciiText left U+%04X in %q", bad, got)
		}
	}
	if want := "接口调用凭证,可使用 access_token (见文档):请勿外传;示例\"小程序\"±"; got != want {
		t.Errorf("asciiText =\n %q\nwant\n %q", got, want)
	}
	// Characters without an ASCII counterpart are untouched.
	if asciiText("错误码。列表、项") != "错误码。列表、项" {
		t.Error("CJK-only punctuation must be preserved")
	}
	if asciiText("") != "" || asciiText("plain ascii") != "plain ascii" {
		t.Error("no-op inputs must be returned unchanged")
	}
}

// Guard: no generated artifact may contain ASCII-confusable punctuation, which
// is what makes editors warn "This document contains many ambiguous unicode
// characters" when opening the contract or the generated Go files.
func TestGeneratedArtifactsAreFreeOfAmbiguousUnicode(t *testing.T) {
	root := filepath.Join("..", "..")
	var files []string
	for _, pat := range []string{
		"miniprogram/*.go", "official/*.go",
		"tools/docsync/swagger/*.json", "tools/docsync/swagger/*.yaml",
	} {
		matches, err := filepath.Glob(filepath.Join(root, pat))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		t.Fatal("no generated artifacts found; run make docs-sync first")
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for i, r := range string(data) {
			ambiguous := (r >= 0xff01 && r <= 0xff5e) ||
				r == '\u00a0' || r == '\u3000' || r == '\u2212' || r == '\u00b7' ||
				r == '\u2018' || r == '\u2019' || r == '\u201c' || r == '\u201d' ||
				r == '\u2013' || r == '\u2014'
			if ambiguous {
				t.Errorf("%s: ambiguous character U+%04X (%q) at offset %d", path, r, r, i)
				break
			}
		}
	}
}
