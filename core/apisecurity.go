package core

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// This file implements the optional "服务通信二次加密和签名" (API security)
// layer: when an account enables API 加密 in the MP backend, the endpoints
// whose documentation declares support must have their request body encrypted
// and signed, and their response signature verified and body decrypted.
//
// Reference:
// https://developers.weixin.qq.com/miniprogram/dev/server/getting_started/api_signature.html
//
// The wire protocol implemented here is the one the guide's own worked example
// uses (endpoint /wxa/getuserriskrank):
//
//	request  body   = {"iv":..,"data":..,"authtag":..}     (AES-256-GCM)
//	request  AAD    = urlpath|appid|timestamp|sn
//	request  plain  = {"_n":..,"_appid":..,"_timestamp":..,<original params>}
//	request  sign   = RSA-PSS-SHA256(salt=32) over
//	                  urlpath \n appid \n timestamp \n <encrypted body>
//	response sign   = RSA-PSS-SHA256 over
//	                  urlpath \n appid \n timestamp \n <encrypted body>
//	response body   = same envelope, decrypted with the same symmetric key
//
// Four documented endpoints that support the layer are GET methods:
//
//	/wxa/business/getuserencryptkey
//	/cgi-bin/express/business/account/getall
//	/cgi-bin/express/business/delivery/getall
//	/cgi-bin/express/business/printer/getall
//
// The guide only ever demonstrates a POST, but it defines the encrypted payload
// as "原请求字段，包含URL参数、POST参数" (the original fields, both URL and POST
// parameters), so for a GET this runtime encrypts the parameters into the body
// as well and keeps access_token in the query. A GET carrying a body is
// unusual, and this path has NOT been verified against the live platform: if
// WeChat rejects it, those four endpoints still work by leaving
// Config.APISecurity nil (they are then called in the clear, as before).
//
// Every other detail is pinned by core/apisecurity_doc_test.go against the
// vectors the guide publishes.

// APIAlg is the request-encryption / signing algorithm pair configured in the
// MP backend under API 安全.
type APIAlg int

const (
	// APIAlgAES256GCM is AES256_GCM request encryption with RSAwithSHA256
	// (RSA-PSS) request signing and platform-certificate response verification.
	APIAlgAES256GCM APIAlg = iota
	// APIAlgSM4GCM is SM4_GCM with SM2withSM3. It is reserved for the
	// documented 国密 configuration but is not implemented yet: unlike the
	// AES/RSA pair, the guide publishes no test vector for it (SM2 signatures
	// are randomised) and no standard-library implementation exists in Go, so
	// the signature encoding could not be verified against WeChat.
	APIAlgSM4GCM
)

// ErrAPISecurityAlgUnsupported reports a configured algorithm this package does
// not implement.
var ErrAPISecurityAlgUnsupported = errors.New("wechat: api security algorithm not supported")

// API security header names, shared by the request and the response.
const (
	headerAppID               = "Wechatmp-AppId"
	headerTimestamp           = "Wechatmp-TimeStamp"
	headerSignature           = "Wechatmp-Signature"
	headerSerial              = "Wechatmp-Serial"
	headerSerialDeprecated    = "Wechatmp-Serial-Deprecated"
	headerSignatureDeprecated = "Wechatmp-Signature-Deprecated"
)

// apiSecurityMaxSkew is how far the response timestamp may differ from the
// caller's clock. The guide's reference implementation rejects responses whose
// _timestamp is more than this far behind, which bounds replay of a captured
// response.
const apiSecurityMaxSkew = 5 * time.Minute

// apiSecurityNonceBytes is the random nonce size for the _n security field. The
// guide recommends 16-32 bytes and its example uses 16 (22 base64 chars).
const apiSecurityNonceBytes = 16

// APISecurityConfig carries the API 安全 keys configured for the account in the
// MP backend, plus the endpoint set the decorator applies to. Build the runtime
// with NewAPISecurity.
type APISecurityConfig struct {
	// SymKey is the API 对称密钥 (base64 of its 32 raw bytes) and SymSN its
	// 密钥编号, used to encrypt the request body and decrypt the response.
	SymKey string
	SymSN  string
	// PrivateKey is the PEM application private key (RSA; PKCS#1 or PKCS#8)
	// used to sign requests, alongside the 平台证书 entries in PlatformCerts
	// (PEM) used to verify responses, keyed by 证书编号.
	PrivateKey    string
	PlatformCerts map[string]string
	// Paths lists the endpoint paths the decorator applies to, e.g.
	// "/wxa/getuserriskrank". Only calls to these paths are encrypted and
	// signed; every other call is left untouched. The caller supplies this set
	// because WeChat documents support per API ("只有部分 API 支持加解密，具体可
	// 参考各 API 文档") and the page for each endpoint carries the
	// "支持加密请求" note. Empty disables the decorator entirely.
	Paths []string
	// Alg selects the algorithm pair; the zero value is APIAlgAES256GCM.
	Alg APIAlg
	// Clock overrides the timestamp source (tests).
	Clock func() time.Time
	// AllowUnsignedResponses accepts a response that carries no
	// Wechatmp-Signature instead of rejecting it. The default (false) is the
	// documented behaviour: an endpoint marked as supporting the layer must
	// return a signed, encrypted response, so a missing signature is treated as
	// a failure. Set it only when a specific deployment is known to receive
	// unsigned replies, because accepting them gives up the tamper protection
	// this layer exists to provide. A body that is shaped like an encrypted
	// envelope is never accepted unsigned, whatever this flag says.
	AllowUnsignedResponses bool
}

// APISecurity is the runtime of the API 二次加密和签名 layer. It implements
// RequestModifier, so a caller installs it on a client with Use and names the
// endpoints it covers through Paths. It is safe for concurrent use.
type APISecurity struct {
	symKey        []byte
	symSN         string
	key           *rsa.PrivateKey
	certs         map[string]*x509.Certificate
	now           func() time.Time
	allowUnsigned bool
	paths         map[string]bool
}

// NewAPISecurity validates cfg and builds the API-security decorator.
func NewAPISecurity(cfg APISecurityConfig) (*APISecurity, error) {
	if cfg.Alg != APIAlgAES256GCM {
		return nil, fmt.Errorf("%w: only AES256_GCM/RSAwithSHA256 is implemented, got %d", ErrAPISecurityAlgUnsupported, cfg.Alg)
	}
	if cfg.SymSN == "" {
		return nil, errors.New("wechat: api security requires the symmetric key number (SymSN)")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.SymKey)
	if err != nil {
		return nil, fmt.Errorf("wechat: decode api security SymKey: %w", err)
	}
	// http.RawStdEncoding because the downloaded key may omit "=" padding.
	if len(key) != 32 {
		if raw, rawErr := base64.RawStdEncoding.DecodeString(strings.TrimRight(cfg.SymKey, "=")); rawErr == nil && len(raw) == 32 {
			key = raw
		} else {
			return nil, fmt.Errorf("wechat: api security SymKey must decode to 32 bytes, got %d", len(key))
		}
	}
	priv, err := parseRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	certs := make(map[string]*x509.Certificate, len(cfg.PlatformCerts))
	for sn, certPEM := range cfg.PlatformCerts {
		cert, certErr := parseCertificate(certPEM)
		if certErr != nil {
			return nil, fmt.Errorf("wechat: platform certificate %q: %w", sn, certErr)
		}
		certs[sn] = cert
	}
	now := cfg.Clock
	if now == nil {
		now = time.Now
	}
	paths := make(map[string]bool, len(cfg.Paths))
	for _, p := range cfg.Paths {
		paths[p] = true
	}
	return &APISecurity{
		symKey:        key,
		symSN:         cfg.SymSN,
		key:           priv,
		certs:         certs,
		now:           now,
		allowUnsigned: cfg.AllowUnsignedResponses,
		paths:         paths,
	}, nil
}

// covers reports whether the decorator applies to an endpoint path.
func (s *APISecurity) covers(path string) bool { return s.paths[path] }

// ModifyRequest implements RequestModifier: for a covered endpoint it encrypts
// the documented parameters into the body and returns the Wechatmp-* signature
// headers. A path outside Paths is left to the runtime.
//
// The parameters come from the typed request struct, and access_token is kept
// out of the payload, as the guide requires. Credentials the plain path injects
// into the query (appid/secret) are injected into the payload here instead.
func (s *APISecurity) ModifyRequest(info RequestInfo, req any) (*ModifiedRequest, error) {
	if !s.covers(info.Path) {
		return nil, nil
	}
	appID := info.Credentials.AppID
	if appID == "" {
		return nil, fmt.Errorf("wechat: api security requires an appid to secure %s", info.Path)
	}
	rv := reflect.Indirect(reflect.ValueOf(req))
	params := injectCredentialParams(collectParams(rv), declaredParams(rv), info.Credentials)
	body, headers, err := s.SecureRequest(appID, info.URLPath, params)
	if err != nil {
		return nil, err
	}
	return &ModifiedRequest{Body: body, Headers: headers}, nil
}

// ModifyResponse implements RequestModifier: for a covered endpoint it verifies
// the platform signature and decrypts the body. An unsigned response is returned
// to the runtime only when AllowUnsignedResponses is set; otherwise it fails the
// call.
func (s *APISecurity) ModifyResponse(info RequestInfo, header http.Header, body []byte) ([]byte, bool, error) {
	if !s.covers(info.Path) {
		return nil, false, nil
	}
	plaintext, err := s.VerifyAndDecryptResponse(info.Credentials.AppID, info.URLPath, header, body)
	if err != nil {
		return nil, false, err
	}
	if plaintext == nil {
		// Unsigned response the caller opted to accept: leave it to the runtime.
		return nil, false, nil
	}
	return plaintext, true, nil
}

// parseRSAPrivateKey accepts a PEM RSA private key in PKCS#1 or PKCS#8 form.
func parseRSAPrivateKey(pemText string) (*rsa.PrivateKey, error) {
	if strings.TrimSpace(pemText) == "" {
		return nil, errors.New("wechat: api security requires an application private key")
	}
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("wechat: application private key is not valid PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("wechat: parse application private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("wechat: application private key is %T, want RSA", parsed)
	}
	return key, nil
}

// parseCertificate accepts a PEM certificate or a bare base64 certificate body.
func parseCertificate(pemText string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("wechat: platform certificate is not valid PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// apiParam is one named value of the encrypted request payload, kept in
// declaration order so the plaintext is byte-for-byte reproducible.
type apiParam struct {
	Name  string
	Value any
}

// collectParams gathers a request struct's documented parameters in field
// order. Both query- and body-documented parameters participate: the API
// security layer carries every parameter except access_token inside the
// encrypted JSON body, which stays in the URL query.
func collectParams(rv reflect.Value) []apiParam {
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return nil
	}
	t := rv.Type()
	var out []apiParam
	for i := range t.NumField() {
		field := t.Field(i)
		fv := rv.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		if tag == "" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if name == "-" || name == "" || name == "access_token" {
			continue
		}
		if omitsZero(tag) && fv.IsZero() {
			continue
		}
		if !fv.CanInterface() {
			continue
		}
		out = append(out, apiParam{Name: name, Value: fv.Interface()})
	}
	return out
}

// SecureRequest encrypts the parameters and signs the ciphertext, returning the
// request body and the Wechatmp-* headers for appID at urlpath.
func (s *APISecurity) SecureRequest(appID, urlpath string, params []apiParam) ([]byte, map[string]string, error) {
	if appID == "" {
		return nil, nil, errors.New("wechat: api security requires an appid")
	}
	ts := s.now().Unix()
	nonce, err := randomBase64(apiSecurityNonceBytes)
	if err != nil {
		return nil, nil, err
	}
	// The security fields lead the payload, matching the guide's reference
	// implementation (Object.assign of the security fields, then the request).
	fields := make([]apiParam, 0, len(params)+3)
	fields = append(fields,
		apiParam{Name: "_n", Value: nonce},
		apiParam{Name: "_appid", Value: appID},
		apiParam{Name: "_timestamp", Value: ts},
	)
	fields = append(fields, params...)
	plaintext, err := marshalOrdered(fields)
	if err != nil {
		return nil, nil, fmt.Errorf("wechat: marshal api security payload: %w", err)
	}
	aad := apiSecurityAAD(urlpath, appID, ts, s.symSN)
	body, err := s.seal(plaintext, aad)
	if err != nil {
		return nil, nil, err
	}
	signature, err := s.signRequest(urlpath, appID, ts, body)
	if err != nil {
		return nil, nil, err
	}
	return body, map[string]string{
		headerAppID:     appID,
		headerTimestamp: strconv.FormatInt(ts, 10),
		headerSignature: signature,
	}, nil
}

// seal encrypts plaintext into the {"iv","data","authtag"} JSON envelope.
func (s *APISecurity) seal(plaintext []byte, aad string) ([]byte, error) {
	block, err := aes.NewCipher(s.symKey)
	if err != nil {
		return nil, fmt.Errorf("wechat: api security cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("wechat: api security gcm: %w", err)
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("wechat: api security iv: %w", err)
	}
	sealed := gcm.Seal(nil, iv, plaintext, []byte(aad))
	split := len(sealed) - gcm.Overhead()
	body, err := marshalOrdered([]apiParam{
		{Name: "iv", Value: base64.StdEncoding.EncodeToString(iv)},
		{Name: "data", Value: base64.StdEncoding.EncodeToString(sealed[:split])},
		{Name: "authtag", Value: base64.StdEncoding.EncodeToString(sealed[split:])},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat: marshal api security envelope: %w", err)
	}
	return body, nil
}

// signRequest computes the base64 RSA-PSS signature over the signed string
// "urlpath\nappid\ntimestamp\nbody". PSS uses SHA-256 with a 32 byte salt, as
// the guide specifies ("签名使用 PSS 填充方式，需要指定 salt 长度为 32").
func (s *APISecurity) signRequest(urlpath, appID string, ts int64, body []byte) (string, error) {
	digest := sha256.Sum256(signedString(urlpath, appID, ts, body))
	sig, err := rsa.SignPSS(rand.Reader, s.key, crypto.SHA256, digest[:], &rsa.PSSOptions{
		SaltLength: 32,
		Hash:       crypto.SHA256,
	})
	if err != nil {
		return "", fmt.Errorf("wechat: sign api request: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyAndDecryptResponse verifies the platform signature of an encrypted
// response and returns the decrypted plaintext. urlpath is the full request
// URL (protocol included, query excluded).
//
// An endpoint that supports the layer must answer with a signed, encrypted
// response, so a missing signature is an error by default. It returns (nil,
// nil) only when the response is unsigned AND the runtime was built with
// AllowUnsignedResponses, which tells the caller to handle the bytes as a
// plain, unverified reply. A body shaped like an encrypted envelope is never
// treated as plain text: that combination would otherwise surface as a
// successful call with zeroed fields.
func (s *APISecurity) VerifyAndDecryptResponse(appID, urlpath string, header http.Header, body []byte) ([]byte, error) {
	timestamp := header.Get(headerTimestamp)
	if header.Get(headerSignature) == "" {
		if looksLikeEnvelope(body) {
			return nil, fmt.Errorf("wechat: response for %s is an encrypted envelope but carries no %s", urlpath, headerSignature)
		}
		if s.allowUnsigned {
			return nil, nil
		}
		return nil, fmt.Errorf("wechat: response for %s carries no %s; set APISecurityConfig.AllowUnsignedResponses to accept unsigned replies", urlpath, headerSignature)
	}
	if timestamp == "" {
		return nil, fmt.Errorf("wechat: encrypted response is missing %s", headerTimestamp)
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("wechat: invalid response %s %q", headerTimestamp, timestamp)
	}
	if err := s.verifyResponse(urlpath, appID, ts, body, header); err != nil {
		return nil, err
	}
	// Only a stale response is rejected: the guide's reference implementation
	// checks `local_ts - resp_ts > 300`, which lets a response whose timestamp
	// is slightly ahead of the local clock pass. Rejecting a future timestamp
	// would fail every call on a host whose clock runs slow.
	if age := s.now().Sub(time.Unix(ts, 0)); age > apiSecurityMaxSkew {
		return nil, fmt.Errorf("wechat: encrypted response is %s old, beyond the %s replay window", age.Round(time.Second), apiSecurityMaxSkew)
	}
	aad := apiSecurityAAD(urlpath, appID, ts, s.symSN)
	plaintext, err := s.open(body, aad)
	if err != nil {
		return nil, err
	}
	if err := validateSecurityFields(plaintext, appID, ts); err != nil {
		return nil, err
	}
	return plaintext, nil
}

// looksLikeEnvelope reports whether body is shaped like the encrypted response
// envelope {"iv":..,"data":..,"authtag":..}. Such a body can never be handled as
// a plain reply: the JSON decoder would drop the envelope fields and hand back a
// zeroed struct with no error.
func looksLikeEnvelope(body []byte) bool {
	var probe struct {
		IV      string `json:"iv"`
		Data    string `json:"data"`
		Authtag string `json:"authtag"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	return probe.IV != "" && probe.Data != "" && probe.Authtag != ""
}

// verifyResponse checks the body signature against the platform certificate
// the response names. During a certificate rollover the response carries both
// the current and the about-to-expire serial/signature pair, so both are tried.
func (s *APISecurity) verifyResponse(urlpath, appID string, ts int64, body []byte, header http.Header) error {
	type candidate struct{ serial, signature string }
	candidates := []candidate{{header.Get(headerSerial), header.Get(headerSignature)}}
	if deprecated := header.Get(headerSerialDeprecated); deprecated != "" {
		candidates = append(candidates, candidate{deprecated, header.Get(headerSignatureDeprecated)})
	}
	var known []string
	for _, c := range candidates {
		if c.serial == "" || c.signature == "" {
			continue
		}
		cert := s.certs[c.serial]
		if cert == nil {
			known = append(known, c.serial)
			continue
		}
		sig, err := base64.StdEncoding.DecodeString(c.signature)
		if err != nil {
			return fmt.Errorf("wechat: decode response signature: %w", err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("wechat: platform certificate %q holds a %T, want RSA", c.serial, cert.PublicKey)
		}
		digest := sha256.Sum256(signedString(urlpath, appID, ts, body))
		// nil options let rsa derive the salt length from the signature.
		if err := rsa.VerifyPSS(pub, crypto.SHA256, digest[:], sig, nil); err == nil {
			return nil
		}
	}
	if len(known) > 0 {
		return fmt.Errorf("wechat: no platform certificate configured for response serial %s", strings.Join(known, ", "))
	}
	return errors.New("wechat: response signature did not verify against any configured platform certificate")
}

// open decrypts a {"iv","data","authtag"} envelope.
func (s *APISecurity) open(body []byte, aad string) ([]byte, error) {
	var env struct {
		IV      string `json:"iv"`
		Data    string `json:"data"`
		Authtag string `json:"authtag"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("wechat: parse encrypted response envelope: %w", err)
	}
	iv, err := base64.StdEncoding.DecodeString(env.IV)
	if err != nil {
		return nil, fmt.Errorf("wechat: decode response iv: %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		return nil, fmt.Errorf("wechat: decode response data: %w", err)
	}
	tag, err := base64.StdEncoding.DecodeString(env.Authtag)
	if err != nil {
		return nil, fmt.Errorf("wechat: decode response authtag: %w", err)
	}
	block, err := aes.NewCipher(s.symKey)
	if err != nil {
		return nil, fmt.Errorf("wechat: api security cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("wechat: api security gcm: %w", err)
	}
	// Guard the nonce: gcm.Open panics on a wrong-size nonce, and a malformed
	// envelope that fails to parse decodes to a zero-length one.
	if len(iv) != gcm.NonceSize() {
		return nil, fmt.Errorf("wechat: encrypted envelope has a %d byte iv, want %d", len(iv), gcm.NonceSize())
	}
	plaintext, err := gcm.Open(nil, iv, append(data, tag...), []byte(aad))
	if err != nil {
		return nil, fmt.Errorf("wechat: decrypt api response: %w", err)
	}
	return plaintext, nil
}

// validateSecurityFields checks the _appid/_timestamp the platform echoed in the
// decrypted body: they must name this app and match the signed header.
func validateSecurityFields(plaintext []byte, appID string, ts int64) error {
	var fields struct {
		AppID     string `json:"_appid"`
		Timestamp int64  `json:"_timestamp"`
	}
	if err := json.Unmarshal(plaintext, &fields); err != nil {
		return fmt.Errorf("wechat: parse decrypted response: %w", err)
	}
	if fields.AppID != appID {
		return fmt.Errorf("wechat: decrypted response _appid %q does not match %q", fields.AppID, appID)
	}
	if fields.Timestamp != ts {
		return fmt.Errorf("wechat: decrypted response _timestamp %d does not match the signed %d", fields.Timestamp, ts)
	}
	return nil
}

// apiSecurityAAD builds the GCM additional authenticated data:
// urlpath|appid|timestamp|sn.
func apiSecurityAAD(urlpath, appID string, ts int64, sn string) string {
	return urlpath + "|" + appID + "|" + strconv.FormatInt(ts, 10) + "|" + sn
}

// signedString builds the signature input: the four fields joined by newlines
// with no trailing separator.
func signedString(urlpath, appID string, ts int64, body []byte) []byte {
	var sb strings.Builder
	sb.Grow(len(urlpath) + len(appID) + len(body) + 32)
	sb.WriteString(urlpath)
	sb.WriteByte('\n')
	sb.WriteString(appID)
	sb.WriteByte('\n')
	sb.WriteString(strconv.FormatInt(ts, 10))
	sb.WriteByte('\n')
	sb.Write(body)
	return []byte(sb.String())
}

// marshalOrdered encodes name/value pairs as a JSON object preserving the given
// order, so the encrypted plaintext is reproducible.
func marshalOrdered(fields []apiParam) ([]byte, error) {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		name, err := json.Marshal(f.Name)
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(f.Value)
		if err != nil {
			return nil, err
		}
		sb.Write(name)
		sb.WriteByte(':')
		sb.Write(value)
	}
	sb.WriteByte('}')
	return []byte(sb.String()), nil
}

// randomBase64 returns n cryptographically random bytes, base64 encoded without
// padding (the format the guide's reference implementation uses for _n and iv).
func randomBase64(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("wechat: generate random bytes: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(raw), nil
}
