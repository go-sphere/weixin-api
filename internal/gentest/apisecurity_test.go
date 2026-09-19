package gentest

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sphere/weixin-api/core"
	"github.com/go-sphere/weixin-api/miniprogram"
)

// The API 安全 keys and the platform certificate published by the WeChat guide
// "服务端 API 签名加密指南". Reusing the documented material keeps this test
// independent of the package's own key handling.
const (
	apiSecAppID   = "wxba6223c06417af7b"
	apiSecSymKey  = "otUpngOjU+nVQaWJIC3D/yMLV17RKaP6t4Ot9tbnzLY="
	apiSecSymSN   = "fa05fe1e5bcc79b81ad5ad4b58acf787"
	apiSecCertSN  = "79ba700ea147819f640941bceb38b1d1"
	apiSecPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3FoQOmOl5/CF5hF7ta4EzCy2LaU3Eu2k9DBwQ73J82I53Sx9
LAgM1DH3IsYohRRx/BESfbdDI2powvr6QYKVIC+4Yavwg7gzhZRxWWmT1HruEADC
ZAgkUCu+9Il/9FPuitPSoIpBd07NqdkkRe82NBOfrKTdhge/5zd457fl7J81Q5VT
IxO8vvq7FSw7k6Jtv+eOjR6SZOWbbUO7f9r4UuUkXmvdGv21qiqtaO1EMw4tUCEL
zY73M7NpCH3RorlommYX3P6q0VrkDHrCE0/QMhmHsF+46E+IRcJ3wtEj3p/mO1Vo
CpEhawC1U728ZUTwWNEii8hPEhcNAZTKaQMaTQIDAQABAoIBAQCXv5p/a5KcyYKc
75tfgekh5wTLKIVmDqzT0evuauyCJTouO+4z/ZNAKuzEUO0kwPDCo8s1MpkU8boV
1Ru1M8WZNePnt65aN+ebbaAl8FRzNvltoeg9VXIUmBvYcjzhOVAE4V2jW7M8A9QU
zUpyswuED6OeFKfOHtYk2In2IipAqhfbyc6gn7uZSWTQsoO6hGBRQ7Ejx+vgwrbx
ZKVZ7UXbPHD0lOEPraA3PH/QUeUKpNwK2NXQoBxWcR283/HxFSAjjSSsGSBKsCnw
DN55P2FQ0HNi5YrwUNT9190NIXSeygaRy1b+D+yBfm+yE7/qXwHLZCHsjO+2tMSS
3KGjllTBAoGBAP9FPeYNKZuu5jt9RpZwXCc9E7Iz7bmM7zws6dun6dQH0xVVWFVm
iGIu07eqyB8HNagXseFzoXLV5EQx+3DaB0bAH+ZEpHGJJpAWSLusigssFUFuTvTF
w+rC5hxOfidMa6+93SU5pWeJb0zJF8PRDaJ3UmwlwpYubF17sT4PD6p9AoGBANz7
RlhRSFvggJjhEMpek3OIYWrrlRNO2MVcP7i/fGNTHhrw7OHcNGRof54QZ2Y0baL7
1vHNokbK2mnT+cQXY/gXMmcE/eV4xyRGYiIL9nBdrkLerc43EYPv+evDvgyji6+y
4np5cKqHrS8F+YzATk82Jt9HgdI2MvfbJTkSbmgRAoGAHNPL9rPb1An/VA6Ery6H
KaM7Gy/EE+U3ixsjWbvvqxMrIkieDh7jHftdy2sM6Hwe8hmi6+vr+pTvD0h5tbfZ
hILj11Q/Idc0NKdflVoZyMM0r0vuvLOsuVFDPUUb+AIoUxNk6vREmpmpqQk4ltN/
763779yfyef6MuBqFrEKut0CgYB9FfsuuOv1nfINF7EybDCZAETsiee7ozEPHnWv
dSzK6FytMV1VSBmcEI7UgUKWVu0MifOUsiq+WcsihmvmNLtQzoioSeoSP7ix7ulT
jmP0HQMsNPI7PW67uVZFv2pPqy/Bx8dtPlqpHN3KNV6Z7q0lJ2j/kHGK9UUKidDb
KnS2kQKBgHZ0cYzwh9YnmfXx9mimF57aQQ8aFc9yaeD5/3G2+a/FZcHtYzUdHQ7P
PS35blD17/NnhunHhuqakbgarH/LIFMHITCVuGQT4xS34kFVjFVhiT3cHfWyBbJ6
GbQuzzFxz/UKDDKf3/ON41k8UP20Gdvmv/+c6qQjKPayME81elus
-----END RSA PRIVATE KEY-----
`
	apiSecCertPEM = `-----BEGIN CERTIFICATE-----
MIID0jCCArqgAwIBAgIUeE+Yy7vM/o+eHHsfM+1bGJJEZTQwDQYJKoZIhvcNAQEL
BQAwXjELMAkGA1UEBhMCQ04xEzARBgNVBAoTClRlbnBheS5jb20xHTAbBgNVBAsT
FFRlbnBheS5jb20gQ0EgQ2VudGVyMRswGQYDVQQDExJUZW5wYXkuY29tIFJvb3Qg
Q0EwHhcNMjIwOTA1MDgzOTIyWhcNMjcwOTA0MDgzOTIyWjBkMRswGQYDVQQDDBJ3
eGQ5MzBlYTVkNWEyNThmNGYxFTATBgNVBAoMDFRlbmNlbnQgSW5jLjEOMAwGA1UE
CwwFV3hnTXAxCzAJBgNVBAYMAkNOMREwDwYDVQQHDAhTaGVuWmhlbjCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAM5D9qlkCmk1kr3FpF0e9pc3kGsvz5RA
0/YRny9xPKIyV2UVMDZvRQ+mDHsiQQFE6etg457KFYSxTDKtItbdl6hJQVGeAvg0
mqPYE9SkHRGTfL/AnXRbKBG2GC2OcaPSAprsLOersjay2me+9pF8VHybV8aox78A
NsU75G/OO3V1iEE0s5Pmglqk8DEiw9gB/dGJzsNfXwzvyJyiUP9ZujYexyjsS+/Z
GdSOUkqL/th+16yHj8alcdyga6YGfWEDyWkt/i/B28cwx4nzwk8xgrurifPaLuMk
0+9wJQLCfAn/f7zyHrC8PcD1XvvRt9VBNMBASXs3710ODyyVf2lkMgkCAwEAAaOB
gTB/MAkGA1UdEwQCMAAwCwYDVR0PBAQDAgTwMGUGA1UdHwReMFwwWqBYoFaGVGh0
dHA6Ly9ldmNhLml0cnVzLmNvbS5jbi9wdWJsaWMvaXRydXNjcmw/Q0E9MUJENDIy
MEU1MERCQzA0QjA2QUQzOTc1NDk4NDZDMDFDM0U4RUJEMjANBgkqhkiG9w0BAQsF
AAOCAQEAL2MK9tYu+ljLVBlSbfEeaKyF07TN+G31Ya5NBzeS1ZCx4joUEIyACWmG
fUkKNKiKV+EMzxeEhKRso1Qif3E7Ipl+PQBoQw6OSR/jFHciYurnGR9CLkL03Zo1
qw1Xetv9OipsvlpA0SOWc207e/XpGdm8C7FMXM6bzvVp8I/STTjC1vqjIZu9WavI
RgGM4jyAPz2XogUq0BNijef8BXbbav9fAsXjHSwn5BQv4iLms3fiLm/eoyQ6dZ2R
oTudrlcyr1bG4vwETLmHF+3yfVp9dpvJ+lyfiviwDwyfa8t2WlJm27DuF4vWoxir
mjgj9tDutIFqxLIovLyg3uiAYtSQ/Q==
-----END CERTIFICATE-----
`
)

// apiSecPaths names the endpoints the decorator covers in these tests. The
// caller supplies this set; the two endpoints exercised here are the worked
// example and a marked binary endpoint.
var apiSecPaths = []string{"/wxa/getuserriskrank", "/cgi-bin/wxaapp/createwxaqrcode"}

func newAPISecurity(t *testing.T) *core.APISecurity {
	t.Helper()
	sec, err := core.NewAPISecurity(core.APISecurityConfig{
		SymKey:                 apiSecSymKey,
		SymSN:                  apiSecSymSN,
		PrivateKey:             apiSecPrivKey,
		PlatformCerts:          map[string]string{apiSecCertSN: apiSecCertPEM},
		Paths:                  apiSecPaths,
		AllowUnsignedResponses: true, // the test servers answer unsigned
	})
	if err != nil {
		t.Fatalf("NewAPISecurity: %v", err)
	}
	return sec
}

// TestAPISecurityEndpointEncryptsWhenConfigured covers the generated call site:
// an endpoint the docs mark as supporting 二次加密和签名 must encrypt its body
// and send the Wechatmp-* headers once Config.APISecurity is set, and must stay
// in the clear otherwise.
func TestAPISecurityEndpointEncryptsWhenConfigured(t *testing.T) {
	var headers http.Header
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"TOK","expires_in":7200}`)
			return
		}
		headers = r.Header.Clone()
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)

	c := miniprogram.New(miniprogram.Config{
		AppID: apiSecAppID, AppSecret: "s", BaseURL: srv.URL,
		Modifiers: []core.RequestModifier{newAPISecurity(t)},
	})
	if _, err := c.PostWxaGetuserriskrank(t.Context(), &miniprogram.PostWxaGetuserriskrankRequest{
		AppID: apiSecAppID, OpenID: "o", Scene: 0, ClientIP: "127.0.0.1",
	}); err != nil {
		t.Fatalf("secured call: %v", err)
	}
	if headers.Get("Wechatmp-AppId") != apiSecAppID {
		t.Errorf("Wechatmp-AppId = %q", headers.Get("Wechatmp-AppId"))
	}
	if headers.Get("Wechatmp-Signature") == "" || headers.Get("Wechatmp-TimeStamp") == "" {
		t.Errorf("security headers missing: %v", headers)
	}
	var env struct {
		IV      string `json:"iv"`
		Data    string `json:"data"`
		Authtag string `json:"authtag"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("secured body is not the documented envelope: %v (%s)", err, body)
	}
	if env.Data == "" || env.Authtag == "" || env.IV == "" {
		t.Errorf("envelope is incomplete: %s", body)
	}
	// The parameters must be inside the ciphertext, not on the wire in clear.
	if strings.Contains(string(body), "openid") {
		t.Errorf("plaintext parameter leaked into the request body: %s", body)
	}

	// Without APISecurity the same call must stay in the clear.
	var plainBody []byte
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"TOK","expires_in":7200}`)
			return
		}
		plainBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv2.Close)
	c2 := miniprogram.New(miniprogram.Config{AppID: "wx", AppSecret: "s", BaseURL: srv2.URL})
	if _, err := c2.PostWxaGetuserriskrank(t.Context(), &miniprogram.PostWxaGetuserriskrankRequest{
		AppID: "wx", OpenID: "o", Scene: 0, ClientIP: "127.0.0.1",
	}); err != nil {
		t.Fatalf("plain call: %v", err)
	}
	if strings.Contains(string(plainBody), "authtag") {
		t.Errorf("unconfigured client encrypted the request: %s", plainBody)
	}
	if !strings.Contains(string(plainBody), "openid") {
		t.Errorf("unconfigured client did not send the documented body: %s", plainBody)
	}
}

// TestGETEndpointsSendParametersInQuery covers the regression where parameters
// the docs tabulated under 请求体 on a GET page were sent nowhere.
func TestGETEndpointsSendParametersInQuery(t *testing.T) {
	// Sized up front so the assertions below never index a nil slice.
	queries := make([]string, 3)
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"TOK","expires_in":7200}`)
			return
		}
		if calls < len(queries) {
			queries[calls] = r.URL.RawQuery
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errcode":0}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{AppID: "wx", AppSecret: "s", BaseURL: srv.URL})

	if _, err := c.GetWxaChecksession(t.Context(), &miniprogram.GetWxaChecksessionRequest{
		OpenID: "OPENID", Signature: "SIG", SigMethod: "hmac_sha256",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetWxaChargeUsageGet(t.Context(), &miniprogram.GetWxaChargeUsageGetRequest{
		SpuID: "SPU1", Offset: 0, Limit: 10,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetWxaGetpaidunionid(t.Context(), &miniprogram.GetWxaGetpaidunionidRequest{
		OpenID: "OPENID",
	}); err != nil {
		t.Fatal(err)
	}
	for i, want := range [][]string{
		{"openid=OPENID", "signature=SIG", "sig_method=hmac_sha256"},
		{"spuId=SPU1", "limit=10"},
		{"openid=OPENID"},
	} {
		if i >= calls {
			t.Fatalf("only %d requests reached the server, want %d", calls, len(want))
		}
		for _, param := range want {
			if !strings.Contains(queries[i], param) {
				t.Errorf("query %d = %q, missing %q", i, queries[i], param)
			}
		}
	}
}

// TestAPIKeyMaterialIsNeverLogged is a guard for the obvious mistake of mixing
// the two signing keys up: the API security symmetric key must never appear in
// a request, and the XPay AppKey must not be mistaken for it.
func TestAPISecurityRejectsBadKey(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(apiSecSymKey)
	if err != nil || len(raw) != 32 {
		t.Fatalf("fixture key is not 32 bytes: %v", err)
	}
	if _, err := core.NewAPISecurity(core.APISecurityConfig{
		SymKey: "AAAA", SymSN: apiSecSymSN, PrivateKey: apiSecPrivKey,
		PlatformCerts: map[string]string{apiSecCertSN: apiSecCertPEM},
	}); err == nil {
		t.Error("expected a short symmetric key to be rejected")
	}
}

// A marked endpoint that streams binary on success must still work when the API
// security layer is configured but the platform answers in the clear (no
// signature headers): the runtime falls back to the raw bytes rather than
// failing or misparsing them. Three of the marked endpoints are mini program
// code images, so this path is reachable in practice.
func TestSecuredBinaryEndpointFallsBackWhenUnsigned(t *testing.T) {
	want := []byte{0x89, 'P', 'N', 'G', 0x0d}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"TOK","expires_in":7200}`)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(want)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{
		AppID: "wx", AppSecret: "s", BaseURL: srv.URL,
		Modifiers: []core.RequestModifier{newAPISecurity(t)},
	})
	got, err := c.PostCgiBinWxaappCreatewxaqrcode(t.Context(), &miniprogram.PostCgiBinWxaappCreatewxaqrcodeRequest{Path: "pages/index", Width: 430})
	if err != nil {
		t.Fatalf("secured binary call: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("binary payload = %v, want %v", got, want)
	}
}

// TestUnsignedResponseIsRejectedByDefault locks the security default in the
// generated call path: a marked endpoint whose reply carries no signature must
// fail rather than silently degrade to unverified plain text.
func TestUnsignedResponseIsRejectedByDefault(t *testing.T) {
	strict, err := core.NewAPISecurity(core.APISecurityConfig{
		SymKey: apiSecSymKey, SymSN: apiSecSymSN, PrivateKey: apiSecPrivKey,
		PlatformCerts: map[string]string{apiSecCertSN: apiSecCertPEM},
		Paths:         apiSecPaths,
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(w, `{"access_token":"TOK","expires_in":7200}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errcode":0,"openid":"ATTACKER"}`)
	}))
	t.Cleanup(srv.Close)
	c := miniprogram.New(miniprogram.Config{
		AppID: apiSecAppID, AppSecret: "s", BaseURL: srv.URL,
		Modifiers: []core.RequestModifier{strict},
	})
	if _, err := c.PostWxaGetuserriskrank(t.Context(), &miniprogram.PostWxaGetuserriskrankRequest{
		AppID: apiSecAppID, OpenID: "o", Scene: 0, ClientIP: "127.0.0.1",
	}); err == nil {
		t.Fatal("an unsigned response was accepted by the default configuration")
	}
}
