package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"slices"
	"strings"
)

// This file is the shared endpoint runtime behind the generated platform
// clients. It owns everything platform-agnostic about issuing a documented
// WeChat API call: tag-driven request assembly, credentials and token
// injection, request signing, and error classification. The generated packages
// contain only type declarations, error-code tables and one-line call sites,
// so this logic exists exactly once instead of once per platform.

// MiniAppEnv is the Mini Program deployment environment selector. It maps onto
// the miniprogram_state value expected by the subscribe-message APIs.
type MiniAppEnv string

const (
	// MiniAppEnvRelease is the released Mini Program version.
	MiniAppEnvRelease MiniAppEnv = "release"
	// MiniAppEnvTrial is the trial (体验版) version.
	MiniAppEnvTrial MiniAppEnv = "trial"
	// MiniAppEnvDevelop is the in-development (开发版) version.
	MiniAppEnvDevelop MiniAppEnv = "develop"
)

func (e MiniAppEnv) String() string { return string(e) }

// SubscribeState maps the environment onto the miniprogram_state value the
// subscribe-message APIs expect: "formal", "trial" or "developer".
func (e MiniAppEnv) SubscribeState() string {
	switch e {
	case MiniAppEnvTrial:
		return "trial"
	case MiniAppEnvDevelop:
		return "developer"
	default:
		return "formal"
	}
}

// SignMode is the request-signing requirement of an endpoint. The generator
// emits the constant as a call argument, so it is compile-time data.
type SignMode int

const (
	// SignNone means the endpoint needs no request signature.
	SignNone SignMode = iota
	// SignPaySig computes pay_sig = HMAC-SHA256(AppKey, path + "&" + body) over
	// the exact bytes sent (the virtual-payment family).
	SignPaySig
	// SignPaySigSession additionally computes
	// signature = HMAC-SHA256(session_key, body) for user-level calls; the
	// request must carry a SessionKey field.
	SignPaySigSession
)

// ErrDoc is the documented meaning of one errcode for a specific endpoint.
type ErrDoc struct {
	Desc     string
	Solution string
}

// Upload is one multipart/form-data file part.
type Upload struct {
	// FieldName is the documented form field of the file part (e.g. "media").
	FieldName string
	// FileName is the name reported to the server.
	FileName string
	// ContentType is optional; it defaults to application/octet-stream.
	ContentType string
	// Reader streams the file content.
	Reader io.Reader
}

// EndpointConfig configures an EndpointClient.
type EndpointConfig struct {
	// AppID / AppSecret fetch and refresh access tokens automatically.
	AppID     string
	AppSecret string
	// Token overrides the AppID/AppSecret flow: when set it is called for every
	// request and no token is cached or refreshed. Use it when the token comes
	// from elsewhere (e.g. a third-party platform).
	Token func(ctx context.Context) (string, error)
	// AppKey is the 米大师 / 虚拟支付 signing key. Endpoints with SignPaySig
	// require it; a signed call fails fast when it is empty.
	AppKey string
	// Env is the Mini Program deployment environment used as the default for
	// miniprogram_state. Empty means MiniAppEnvRelease.
	Env MiniAppEnv
	// Cache stores access tokens; nil creates an in-process memory cache.
	Cache Cache
	// Proxy is an optional HTTP(S) proxy URL for outbound requests.
	Proxy string
	// BaseURL overrides the API host (tests, gateways).
	BaseURL string
	// HTTPClient injects a custom HTTP client.
	HTTPClient *http.Client
	// Modifiers decorate outgoing requests and incoming responses. Use installs
	// more on an existing client. nil means no decoration.
	Modifiers []RequestModifier
}

// EndpointClient issues documented WeChat API calls. Token caching,
// de-duplicated refresh and the retry after a token failure come from the
// embedded Client.
type EndpointClient struct {
	*Client
	tokenFn   func(ctx context.Context) (string, error)
	appID     string
	appKey    string
	env       MiniAppEnv
	modifiers []RequestModifier
}

// NewEndpointClient builds a client for one WeChat platform account.
func NewEndpointClient(cfg EndpointConfig) *EndpointClient {
	env := cfg.Env
	if env == "" {
		env = MiniAppEnvRelease
	}
	return &EndpointClient{
		Client: NewClient(CredentialsOf(Credentials{
			AppID:     cfg.AppID,
			AppSecret: cfg.AppSecret,
		}), ClientOptions{
			Cache:      cfg.Cache,
			Proxy:      cfg.Proxy,
			BaseURL:    cfg.BaseURL,
			HTTPClient: cfg.HTTPClient,
		}),
		tokenFn:   cfg.Token,
		appID:     cfg.AppID,
		appKey:    cfg.AppKey,
		env:       env,
		modifiers: slices.Clone(cfg.Modifiers),
	}
}

// WithCredentials derives a client for another account, sharing the token cache,
// HTTP client and request modifiers so one process can serve several accounts.
func (c *EndpointClient) WithCredentials(appID, appSecret string) *EndpointClient {
	return NewEndpointClient(EndpointConfig{
		AppID:      appID,
		AppSecret:  appSecret,
		AppKey:     c.appKey,
		Env:        c.env,
		Token:      c.tokenFn,
		Cache:      c.cache,
		BaseURL:    c.baseURL,
		HTTPClient: c.httpClient,
		Modifiers:  c.modifiers,
	})
}

// Env returns the effective deployment environment.
func (c *EndpointClient) Env() MiniAppEnv { return c.env }

// AccessToken returns a usable access token. A caller-supplied Token function
// takes precedence and disables caching.
func (c *EndpointClient) AccessToken(ctx context.Context, reload bool) (string, error) {
	if c.tokenFn != nil {
		return c.tokenFn(ctx)
	}
	return c.GetAccessToken(ctx, reload)
}

// GetJsSDKConfig builds the signed wx.config payload for a page URL served
// under the account, fetching a jsapi ticket first. The signature algorithm is
// shared with the rest of the library through this package.
func (c *EndpointClient) GetJsSDKConfig(ctx context.Context, pageURL string) (*JSSDKConfig, error) {
	ticket, err := c.jsapiTicket(ctx)
	if err != nil {
		return nil, err
	}
	return NewJSSDKConfig(c.appID, ticket, pageURL)
}

// jsapiTicket fetches the JS-SDK ticket (type=jsapi).
func (c *EndpointClient) jsapiTicket(ctx context.Context) (string, error) {
	resp, err := Call[jsapiTicketResponse](c, ctx, http.MethodGet, "/cgi-bin/ticket/getticket",
		&jsapiTicketRequest{Type: "jsapi"}, nil, true, SignNone)
	if err != nil {
		return "", err
	}
	if resp.Ticket == "" {
		return "", errors.New("wechat: empty jsapi ticket in response")
	}
	return resp.Ticket, nil
}

type jsapiTicketRequest struct {
	Type string `query:"type"`
}

type jsapiTicketResponse struct {
	Ticket string `json:"ticket"`
}

// omitsZero reports whether tag carries an option that drops the field when it
// holds the zero value. The generator emits omitzero for numeric and boolean
// fields (their zero value means "absent") and omitempty for strings, slices
// and maps; both spellings mean the same thing to this runtime.
func omitsZero(tag string) bool {
	_, opts, ok := strings.Cut(tag, ",")
	if !ok {
		return false
	}
	for opt := range strings.SplitSeq(opts, ",") {
		switch strings.TrimSpace(opt) {
		case "omitempty", "omitzero":
			return true
		}
	}
	return false
}

// setQueryValue writes one query-tagged field. A slice is written as one
// name=value pair per element, the conventional multi-value encoding; the other
// kinds use their default formatting. (No endpoint in the generated clients
// carries a slice query parameter today, so this path is defensive.)
func setQueryValue(values url.Values, name string, fv reflect.Value) {
	if fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array {
		for i := range fv.Len() {
			values.Add(name, fmt.Sprintf("%v", fv.Index(i).Interface()))
		}
		return
	}
	values.Set(name, fmt.Sprintf("%v", fv.Interface()))
}

// applyTags fills the URL query and JSON body from a request struct's tags.
// Fields documented as required carry no omitempty and are always sent, even
// when they hold the zero value. declared receives every query key the struct
// defines, so the caller can fill in credentials the request left empty.
func applyTags(rv reflect.Value, values url.Values, declared map[string]bool) (map[string]any, error) {
	body := map[string]any{}
	t := rv.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fv := rv.Field(i)
		if q := field.Tag.Get("query"); q != "" {
			name, _, _ := strings.Cut(q, ",")
			if name == "-" || name == "" {
				continue
			}
			if declared != nil {
				declared[name] = true
			}
			// Ownership of access_token/pay_sig is decided at generation time:
			// the generator omits those fields where this runtime supplies them,
			// so any field present here is caller-supplied data (the /sns/ OAuth
			// token, a B2B pay_sig) and must be sent as written.
			if omitsZero(q) && fv.IsZero() {
				continue
			}
			setQueryValue(values, name, fv)
			continue
		}
		j := field.Tag.Get("json")
		if j == "" || j == "-" {
			continue
		}
		name, _, _ := strings.Cut(j, ",")
		if omitsZero(j) && fv.IsZero() {
			continue
		}
		body[name] = fv.Interface()
	}
	if len(body) == 0 {
		return nil, nil
	}
	return body, nil
}

// injectCredentials supplies appid/secret for the endpoints that authenticate
// with the application credentials instead of an access token
// (/sns/jscode2session, /sns/oauth2/...). Only parameters the request struct
// already declares are touched, so no endpoint gains an unexpected parameter.
func (c *EndpointClient) injectCredentials(values url.Values, declared map[string]bool) {
	creds := c.Credentials()
	if declared["appid"] && values.Get("appid") == "" && creds.AppID != "" {
		values.Set("appid", creds.AppID)
	}
	if declared["secret"] && values.Get("secret") == "" && creds.AppSecret != "" {
		values.Set("secret", creds.AppSecret)
	}
}

// buildRequest assembles a JSON request: query-tagged fields become URL
// parameters, json-tagged fields the request body, and the access token plus
// any required signature are appended.
//
// Before assembling, the installed RequestModifiers are asked for a decorated
// request (see modifier.go); one that accepts completely replaces the body and
// adds its headers, which is how the API security layer encrypts and signs a
// call without the generated code knowing about it.
func (c *EndpointClient) buildRequest(ctx context.Context, token, method, path string, req any, sign SignMode) (*http.Request, error) {
	values := url.Values{}
	if token != "" {
		values.Set("access_token", token)
	}
	var body []byte
	declared := map[string]bool{}
	var sessionKey string
	if req != nil {
		rv := reflect.Indirect(reflect.ValueOf(req))
		if rv.Kind() == reflect.Struct {
			bodyMap, err := applyTags(rv, values, declared)
			if err != nil {
				return nil, err
			}
			if method != http.MethodGet && len(bodyMap) > 0 {
				buf, err := json.Marshal(bodyMap)
				if err != nil {
					return nil, err
				}
				body = buf
			}
			if sign == SignPaySigSession {
				sessionKey = sessionKeyOf(rv)
				if sessionKey == "" {
					return nil, fmt.Errorf("wechat: %s requires a SessionKey for its user-level signature", path)
				}
			}
		}
	}
	c.injectCredentials(values, declared)
	// Sign the exact bytes that go on the wire.
	if sign != SignNone {
		if c.appKey == "" {
			return nil, fmt.Errorf("wechat: %s requires a non-empty AppKey to compute pay_sig", path)
		}
		values.Set("pay_sig", hmacSHA256Hex([]byte(c.appKey), []byte(path+"&"+string(body))))
	}
	if sign == SignPaySigSession {
		values.Set("signature", hmacSHA256Hex([]byte(sessionKey), body))
	}
	// Give the installed decorators their chance to replace the body and add
	// headers. They see the typed request, so a parameter's declared Go type is
	// preserved.
	var extraHeaders map[string]string
	info := RequestInfo{URLPath: c.baseURL + path, Path: path, Method: method, Credentials: c.Credentials()}
	if modified, err := c.modifyRequest(info, req); err != nil {
		return nil, err
	} else if modified != nil {
		body = modified.Body
		extraHeaders = modified.Headers
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path+"?"+values.Encode(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// WeChat expects a JSON content type on write methods even when the body
	// is empty, so it is set unconditionally for non-GET calls.
	if method != http.MethodGet {
		httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	for name, value := range extraHeaders {
		httpReq.Header.Set(name, value)
	}
	return httpReq, nil
}

// declaredParams returns the set of documented parameter names a request struct
// defines (both query- and body-tagged), mirroring the declared map applyTags
// fills.
func declaredParams(rv reflect.Value) map[string]bool {
	declared := map[string]bool{}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return declared
	}
	t := rv.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		for _, key := range []string{"query", "json"} {
			tag := field.Tag.Get(key)
			if tag == "" || tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			if name != "" && name != "-" {
				declared[name] = true
			}
		}
	}
	return declared
}

// injectCredentialParams fills appid/secret for the endpoints that authenticate
// with the application credentials instead of an access token, matching
// injectCredentials on the plain path: only parameters the struct declares are
// touched, and a caller-supplied non-empty value wins.
func injectCredentialParams(params []apiParam, declared map[string]bool, creds Credentials) []apiParam {
	fill := func(name, value string) {
		if !declared[name] || value == "" {
			return
		}
		for i := range params {
			if params[i].Name == name {
				if s, ok := params[i].Value.(string); ok && s == "" {
					params[i].Value = value
				}
				return
			}
		}
		params = append(params, apiParam{Name: name, Value: value})
	}
	fill("appid", creds.AppID)
	fill("secret", creds.AppSecret)
	return params
}

// sessionKeyOf reads the synthetic SessionKey field the generator adds to the
// request structs of user-signed endpoints.
func sessionKeyOf(rv reflect.Value) string {
	f := rv.FieldByName("SessionKey")
	if f.IsValid() && f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}

// buildUploadRequest assembles a multipart/form-data request from the same
// request struct plus one file part.
func (c *EndpointClient) buildUploadRequest(ctx context.Context, token, path string, req any, file *Upload) (*http.Request, error) {
	values := url.Values{}
	if token != "" {
		values.Set("access_token", token)
	}
	var form map[string]any
	declared := map[string]bool{}
	if req != nil {
		rv := reflect.Indirect(reflect.ValueOf(req))
		if rv.Kind() == reflect.Struct {
			var err error
			form, err = applyTags(rv, values, declared)
			if err != nil {
				return nil, err
			}
		}
	}
	c.injectCredentials(values, declared)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range form {
		if err := mw.WriteField(k, fmt.Sprintf("%v", v)); err != nil {
			return nil, err
		}
	}
	if file != nil && file.Reader != nil {
		name := file.FieldName
		if name == "" {
			name = "media"
		}
		contentType := file.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, name, file.FileName))
		h.Set("Content-Type", contentType)
		part, err := mw.CreatePart(h)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, file.Reader); err != nil {
			return nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+"?"+values.Encode(), &buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	return httpReq, nil
}

// do sends the request and returns the response bytes, turning a non-zero
// errcode envelope into a classified error.
//
// The installed RequestModifiers may replace the response body before the errcode
// envelope is inspected (see modifier.go), which is how the API security layer
// verifies and decrypts an encrypted reply.
func (c *EndpointClient) do(ctx context.Context, httpReq *http.Request, path string) ([]byte, error) {
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat: unexpected HTTP status %d for %s %s", resp.StatusCode, httpReq.Method, httpReq.URL.Path)
	}
	info := RequestInfo{URLPath: c.baseURL + path, Path: path, Method: httpReq.Method, Credentials: c.Credentials()}
	if data, err = c.modifyResponse(info, resp.Header, data); err != nil {
		return nil, err
	}
	if bizErr := scanErrorBody(data); bizErr != nil {
		return nil, bizErr
	}
	return data, nil
}

// attachErrDoc copies the endpoint's documented meaning of the errcode onto the
// error.
func attachErrDoc(err error, errDocs map[int]ErrDoc) {
	if err == nil || len(errDocs) == 0 {
		return
	}
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		if doc, found := errDocs[apiErr.ErrCode]; found {
			apiErr.Description = doc.Desc
			apiErr.Solution = doc.Solution
			apiErr.Doc = new(doc)
		}
	}
}

// send runs one request with token handling: it obtains the token, performs the
// call and retries once with a refreshed token when WeChat reports an
// access-token failure.
func (c *EndpointClient) send(ctx context.Context, method, path string, auth bool, build func(token string) (*http.Request, error), errDocs map[int]ErrDoc) ([]byte, error) {
	var token string
	if auth {
		t, tokenErr := c.AccessToken(ctx, false)
		if tokenErr != nil {
			return nil, tokenErr
		}
		token = t
	}
	httpReq, err := build(token)
	if err != nil {
		return nil, err
	}
	data, err := c.do(ctx, httpReq, path)
	if err == nil || !auth || c.tokenFn != nil || !isNeedRetryError(err) {
		attachErrDoc(err, errDocs)
		return data, err
	}
	// The token was rejected: force a refresh and retry once.
	token, refreshErr := c.AccessToken(ctx, true)
	if refreshErr != nil {
		return nil, refreshErr
	}
	httpReq, err = build(token)
	if err != nil {
		return nil, err
	}
	data, err = c.do(ctx, httpReq, path)
	attachErrDoc(err, errDocs)
	return data, err
}

// Call executes one documented endpoint that returns a JSON payload. auth
// reports whether the endpoint takes an access_token; sign is its signing
// requirement.
func Call[T any](c *EndpointClient, ctx context.Context, method, path string, req any, errDocs map[int]ErrDoc, auth bool, sign SignMode) (*T, error) {
	data, err := c.send(ctx, method, path, auth, func(token string) (*http.Request, error) {
		return c.buildRequest(ctx, token, method, path, req, sign)
	}, errDocs)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("wechat: decode %s %s: %w", method, path, err)
	}
	return &out, nil
}

// CallBinary executes one documented endpoint that returns raw bytes on success
// (media downloads, mini program code images) and the JSON errcode envelope
// only on failure.
func CallBinary(c *EndpointClient, ctx context.Context, method, path string, req any, errDocs map[int]ErrDoc, auth bool, sign SignMode) ([]byte, error) {
	return c.send(ctx, method, path, auth, func(token string) (*http.Request, error) {
		return c.buildRequest(ctx, token, method, path, req, sign)
	}, errDocs)
}

// CallUpload executes one documented multipart/form-data upload endpoint.
//
// Uploads are not decorated: the documented API 二次加密 does not cover them,
// so the body stays the multipart form the endpoint requires.
func CallUpload(c *EndpointClient, ctx context.Context, path string, req any, file *Upload, errDocs map[int]ErrDoc, auth bool) ([]byte, error) {
	return c.send(ctx, http.MethodPost, path, auth, func(token string) (*http.Request, error) {
		return c.buildUploadRequest(ctx, token, path, req, file)
	}, errDocs)
}

// Decode converts a successful raw response into the typed payload.
func Decode[T any](ctx context.Context, method, path string, data []byte, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	var out T
	if jsonErr := json.Unmarshal(data, &out); jsonErr != nil {
		return nil, fmt.Errorf("wechat: decode %s %s: %w", method, path, jsonErr)
	}
	return &out, nil
}
