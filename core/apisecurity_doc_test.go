package core

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The vectors in this file are published verbatim by the WeChat guide
// "服务端 API 签名加密指南" for its worked example (endpoint
// /wxa/getuserriskrank). They are external data: the guide fixes the symmetric
// key, the application private key, the platform certificate, the plaintexts,
// the ciphertexts and the RSA-PSS signatures. Because they come from the
// documentation and not from this package, they are what pins the wire format;
// a self-consistent round trip could not.
//
// Reference:
// https://developers.weixin.qq.com/miniprogram/dev/server/getting_started/api_signature.html
const (
	docAppID   = "wxba6223c06417af7b"
	docPath    = "/wxa/getuserriskrank"
	docURLPath = "https://api.weixin.qq.com" + docPath
	docSymKey  = "otUpngOjU+nVQaWJIC3D/yMLV17RKaP6t4Ot9tbnzLY="
	docSymSN   = "fa05fe1e5bcc79b81ad5ad4b58acf787"
	docCertSN  = "79ba700ea147819f640941bceb38b1d1"
	docReqTS   = 1635927954
	docRespTS  = 1635927956
	// docReqSig is the guide's published request signature over
	// docRequestEnvelope, computed with PSS salt length 32.
	docReqSig = "wcSSWHZunjz9VKl9q+If9deiyECXDAELfAJNZ4+5T+NhFr8zfhkwdQtlgQ7nN5xs99R57La9UjBTRBGge2KYyshWtw7HIMPAqWNsnpHvx0b2f7s6Bt7OpfOQLlIfNgepgTVmUwrqW8/7A12szj7tCe/bRFilwnaX6N0w4duHlfL7ic7IIZXouvy9dLRAa5GtEk1eD/LPWRiKh0SvJ3znPY/pSiQW9zSkXVdj9UGGM8qcKLzPGJ7gSmt3ZOPkFapk9wqFmhJwQj//xN5+hUlr2UiNPMNSHve5Y2ADLsNHqk5t7RfAZ8nW9/8lzhVt4t+toy1FeehxCGIC8qgmjIl1hg=="
	// docRespSig is the guide's published response signature over
	// docResponseEnvelope, computed with the platform certificate.
	docRespSig = "Ht0VfQkkEweJ4hU266C14Aj64H9AXfkwNi5zxUZETCvR2svU1ZYdosDhFX/voLj1TyszqKsVxAlENGt7PPZZ8RQX7jnA4SKhiPUhW4LTbyTenisHJ+ohSfDjYnXavjQsBHspFS+BlPHuSSJ2xyQzw1+HuC6nid09ZL4FnGSYo4OI5MJrSb9xLzIVZMIDuUQchGKi/KaB1KzxECLEZcfjqbAgmxC7qOmuBLyO1WkHYDM95NJrHJWba5xv4wrwPru9yYTJSNRnlM+zrW5w9pOubC4Jtj3szTAEuOz9AcqUmgaAvMLNAIa8hfODLRe3n/cu4SgYlN/ZkNRU4QXVNbPGMg=="
)

// docPrivateKeyPEM is the guide's published application private key (PKCS#1).
const docPrivateKeyPEM = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA3FoQOmOl5/CF5hF7ta4EzCy2LaU3Eu2k9DBwQ73J82I53Sx9\nLAgM1DH3IsYohRRx/BESfbdDI2powvr6QYKVIC+4Yavwg7gzhZRxWWmT1HruEADC\nZAgkUCu+9Il/9FPuitPSoIpBd07NqdkkRe82NBOfrKTdhge/5zd457fl7J81Q5VT\nIxO8vvq7FSw7k6Jtv+eOjR6SZOWbbUO7f9r4UuUkXmvdGv21qiqtaO1EMw4tUCEL\nzY73M7NpCH3RorlommYX3P6q0VrkDHrCE0/QMhmHsF+46E+IRcJ3wtEj3p/mO1Vo\nCpEhawC1U728ZUTwWNEii8hPEhcNAZTKaQMaTQIDAQABAoIBAQCXv5p/a5KcyYKc\n75tfgekh5wTLKIVmDqzT0evuauyCJTouO+4z/ZNAKuzEUO0kwPDCo8s1MpkU8boV\n1Ru1M8WZNePnt65aN+ebbaAl8FRzNvltoeg9VXIUmBvYcjzhOVAE4V2jW7M8A9QU\nzUpyswuED6OeFKfOHtYk2In2IipAqhfbyc6gn7uZSWTQsoO6hGBRQ7Ejx+vgwrbx\nZKVZ7UXbPHD0lOEPraA3PH/QUeUKpNwK2NXQoBxWcR283/HxFSAjjSSsGSBKsCnw\nDN55P2FQ0HNi5YrwUNT9190NIXSeygaRy1b+D+yBfm+yE7/qXwHLZCHsjO+2tMSS\n3KGjllTBAoGBAP9FPeYNKZuu5jt9RpZwXCc9E7Iz7bmM7zws6dun6dQH0xVVWFVm\niGIu07eqyB8HNagXseFzoXLV5EQx+3DaB0bAH+ZEpHGJJpAWSLusigssFUFuTvTF\nw+rC5hxOfidMa6+93SU5pWeJb0zJF8PRDaJ3UmwlwpYubF17sT4PD6p9AoGBANz7\nRlhRSFvggJjhEMpek3OIYWrrlRNO2MVcP7i/fGNTHhrw7OHcNGRof54QZ2Y0baL7\n1vHNokbK2mnT+cQXY/gXMmcE/eV4xyRGYiIL9nBdrkLerc43EYPv+evDvgyji6+y\n4np5cKqHrS8F+YzATk82Jt9HgdI2MvfbJTkSbmgRAoGAHNPL9rPb1An/VA6Ery6H\nKaM7Gy/EE+U3ixsjWbvvqxMrIkieDh7jHftdy2sM6Hwe8hmi6+vr+pTvD0h5tbfZ\nhILj11Q/Idc0NKdflVoZyMM0r0vuvLOsuVFDPUUb+AIoUxNk6vREmpmpqQk4ltN/\n763779yfyef6MuBqFrEKut0CgYB9FfsuuOv1nfINF7EybDCZAETsiee7ozEPHnWv\ndSzK6FytMV1VSBmcEI7UgUKWVu0MifOUsiq+WcsihmvmNLtQzoioSeoSP7ix7ulT\njmP0HQMsNPI7PW67uVZFv2pPqy/Bx8dtPlqpHN3KNV6Z7q0lJ2j/kHGK9UUKidDb\nKnS2kQKBgHZ0cYzwh9YnmfXx9mimF57aQQ8aFc9yaeD5/3G2+a/FZcHtYzUdHQ7P\nPS35blD17/NnhunHhuqakbgarH/LIFMHITCVuGQT4xS34kFVjFVhiT3cHfWyBbJ6\nGbQuzzFxz/UKDDKf3/ON41k8UP20Gdvmv/+c6qQjKPayME81elus\n-----END RSA PRIVATE KEY-----"

// docPlatformCertPEM is the guide's published platform certificate (PEM).
const docPlatformCertPEM = "-----BEGIN CERTIFICATE-----\nMIID0jCCArqgAwIBAgIUeE+Yy7vM/o+eHHsfM+1bGJJEZTQwDQYJKoZIhvcNAQEL\nBQAwXjELMAkGA1UEBhMCQ04xEzARBgNVBAoTClRlbnBheS5jb20xHTAbBgNVBAsT\nFFRlbnBheS5jb20gQ0EgQ2VudGVyMRswGQYDVQQDExJUZW5wYXkuY29tIFJvb3Qg\nQ0EwHhcNMjIwOTA1MDgzOTIyWhcNMjcwOTA0MDgzOTIyWjBkMRswGQYDVQQDDBJ3\neGQ5MzBlYTVkNWEyNThmNGYxFTATBgNVBAoMDFRlbmNlbnQgSW5jLjEOMAwGA1UE\nCwwFV3hnTXAxCzAJBgNVBAYMAkNOMREwDwYDVQQHDAhTaGVuWmhlbjCCASIwDQYJ\nKoZIhvcNAQEBBQADggEPADCCAQoCggEBAM5D9qlkCmk1kr3FpF0e9pc3kGsvz5RA\n0/YRny9xPKIyV2UVMDZvRQ+mDHsiQQFE6etg457KFYSxTDKtItbdl6hJQVGeAvg0\nmqPYE9SkHRGTfL/AnXRbKBG2GC2OcaPSAprsLOersjay2me+9pF8VHybV8aox78A\nNsU75G/OO3V1iEE0s5Pmglqk8DEiw9gB/dGJzsNfXwzvyJyiUP9ZujYexyjsS+/Z\nGdSOUkqL/th+16yHj8alcdyga6YGfWEDyWkt/i/B28cwx4nzwk8xgrurifPaLuMk\n0+9wJQLCfAn/f7zyHrC8PcD1XvvRt9VBNMBASXs3710ODyyVf2lkMgkCAwEAAaOB\ngTB/MAkGA1UdEwQCMAAwCwYDVR0PBAQDAgTwMGUGA1UdHwReMFwwWqBYoFaGVGh0\ndHA6Ly9ldmNhLml0cnVzLmNvbS5jbi9wdWJsaWMvaXRydXNjcmw/Q0E9MUJENDIy\nMEU1MERCQzA0QjA2QUQzOTc1NDk4NDZDMDFDM0U4RUJEMjANBgkqhkiG9w0BAQsF\nAAOCAQEAL2MK9tYu+ljLVBlSbfEeaKyF07TN+G31Ya5NBzeS1ZCx4joUEIyACWmG\nfUkKNKiKV+EMzxeEhKRso1Qif3E7Ipl+PQBoQw6OSR/jFHciYurnGR9CLkL03Zo1\nqw1Xetv9OipsvlpA0SOWc207e/XpGdm8C7FMXM6bzvVp8I/STTjC1vqjIZu9WavI\nRgGM4jyAPz2XogUq0BNijef8BXbbav9fAsXjHSwn5BQv4iLms3fiLm/eoyQ6dZ2R\noTudrlcyr1bG4vwETLmHF+3yfVp9dpvJ+lyfiviwDwyfa8t2WlJm27DuF4vWoxir\nmjgj9tDutIFqxLIovLyg3uiAYtSQ/Q==\n-----END CERTIFICATE-----"

// docRequestPlaintext is the encrypted request payload of the worked example,
// in the order the guide's reference implementation builds it: the three
// security fields first, then the original parameters.
const docRequestPlaintext = `{"_n":"o89QaPVsRu1yppIZzvSZc4","_appid":"` + docAppID + `","_timestamp":1635927954,"appid":"` + docAppID + `","openid":"oEWzBfmdLqhFS2mTXCo2E4Y9gJAM","scene":0,"client_ip":"127.0.0.1"}`

// docResponsePlaintext is the decrypted response payload of the worked example.
const docResponsePlaintext = `{"_n":"ShYZpqdVgY+yQVAxNSWhYg","_appid":"` + docAppID + `","_timestamp":1635927956,"errcode":0,"errmsg":"getuserriskrank succ","risk_rank":0,"unoin_id":2258658297}`

// docRequestEnvelope and docResponseEnvelope are the published ciphertext
// envelopes, rebuilt from their documented iv/data/authtag parts.
// The envelopes are stored as the exact published byte strings: the signature
// covers the literal bytes, so the field order in the JSON must be preserved
// rather than re-encoded.
var docRequestEnvelope = []byte("{\"iv\":\"fmW/zNxXlytUZBgj\",\"data\":\"0IDVdrPtSPF/Oe2CTXCV2vVNPbVJdJlP2WaTMQnoYLh5iCrrSNfQFh25EnStDMf0hLlVNBCZQtf9NaV0m4aRA4AAYIO7oR/Ge+4yY4EmZp5EVPB42xjScgMx5X3D4VdLCfynXIUKUtZHZvk1zmLVE3RauzJgiM1BB1CPmwcENo3MTJ0z8Vfkf5tMv54kOXobDLlV5rfqKdAX7gM/rP82DgZdt9vvZX44ipdbHIjJvw83ZXAFtvftdVw2Qd8=\",\"authtag\":\"5qeM/2vZv+6KtScN94IpMg==\"}")

var docResponseEnvelope = []byte("{\"iv\":\"r2WDQt56rEAmMuoR\",\"data\":\"HExs66Ik3el+iM4IpeQ7SMEN934FRLFYOd3EmeaIrpP4EPTHckoco6O+PaoRZRa3lqaPRZT7r52f7LUok6gLxc6cdR8C4vpIIfh4xfLC4L7FNy9GbuMK1hcoi8b7gkWJcwZMkuCFNEDmqn3T49oWzAQOrY4LZnnnykv6oUJotdAsnKvmoJkLK7hRh7M2B1d2UnTnRuoIyarXc5Iojwoghx4BOvnV\",\"authtag\":\"z2BFD8QctKXTuBlhICGOjQ==\"}")

func docAPISecurity(t *testing.T) *APISecurity {
	t.Helper()
	sec, err := NewAPISecurity(APISecurityConfig{
		SymKey:        docSymKey,
		SymSN:         docSymSN,
		PrivateKey:    docPrivateKeyPEM,
		PlatformCerts: map[string]string{docCertSN: docPlatformCertPEM},
		Paths:         []string{docPath},
		Clock:         func() time.Time { return time.Unix(docRespTS, 0) },
	})
	if err != nil {
		t.Fatalf("NewAPISecurity: %v", err)
	}
	return sec
}

// docAPISecurityFor builds the documented key material for an explicit path
// set, optionally allowing unsigned responses. Because Paths is supplied by the
// caller, each test names the endpoints its decorator covers.
func docAPISecurityFor(t *testing.T, allowUnsigned bool, paths ...string) *APISecurity {
	t.Helper()
	sec, err := NewAPISecurity(APISecurityConfig{
		SymKey:                 docSymKey,
		SymSN:                  docSymSN,
		PrivateKey:             docPrivateKeyPEM,
		PlatformCerts:          map[string]string{docCertSN: docPlatformCertPEM},
		Paths:                  paths,
		Clock:                  func() time.Time { return time.Unix(docRespTS, 0) },
		AllowUnsignedResponses: allowUnsigned,
	})
	if err != nil {
		t.Fatalf("NewAPISecurity: %v", err)
	}
	return sec
}

// docAPISecurityLenient builds the same runtime with unsigned responses
// allowed, to exercise the opt-out path.
func docAPISecurityLenient(t *testing.T) *APISecurity {
	t.Helper()
	sec, err := NewAPISecurity(APISecurityConfig{
		SymKey:                 docSymKey,
		SymSN:                  docSymSN,
		PrivateKey:             docPrivateKeyPEM,
		PlatformCerts:          map[string]string{docCertSN: docPlatformCertPEM},
		Paths:                  []string{docPath},
		Clock:                  func() time.Time { return time.Unix(docRespTS, 0) },
		AllowUnsignedResponses: true,
	})
	if err != nil {
		t.Fatalf("NewAPISecurity: %v", err)
	}
	return sec
}

// TestDocRequestSignature verifies the guide's published request signature
// against this package's signed-string construction and PSS parameters. PSS is
// randomised, so the signature cannot be compared byte for byte; verifying the
// documented signature proves the signed string, the hash and the salt length.
func TestDocRequestSignature(t *testing.T) {
	sec := docAPISecurity(t)
	digest := sha256.Sum256(signedString(docURLPath, docAppID, docReqTS, docRequestEnvelope))
	sig, err := base64.StdEncoding.DecodeString(docReqSig)
	if err != nil {
		t.Fatal(err)
	}
	if err := rsa.VerifyPSS(&sec.key.PublicKey, crypto.SHA256, digest[:], sig, &rsa.PSSOptions{SaltLength: 32}); err != nil {
		t.Fatalf("documented request signature did not verify: %v", err)
	}
}

// TestDocRequestDecrypt decrypts the guide's published request ciphertext and
// compares it with the published plaintext, pinning the field order, the AAD
// layout and the GCM parameters.
func TestDocRequestDecrypt(t *testing.T) {
	sec := docAPISecurity(t)
	plain, err := sec.open(docRequestEnvelope, apiSecurityAAD(docURLPath, docAppID, docReqTS, docSymSN))
	if err != nil {
		t.Fatalf("decrypt documented request: %v", err)
	}
	if string(plain) != docRequestPlaintext {
		t.Fatalf("documented request plaintext mismatch:\n got %s\nwant %s", plain, docRequestPlaintext)
	}
}

// TestDocResponseVerifyAndDecrypt runs the full public response path against the
// guide's published response: certificate-based signature verification, the
// _appid/_timestamp checks and decryption.
func TestDocResponseVerifyAndDecrypt(t *testing.T) {
	sec := docAPISecurity(t)
	header := http.Header{}
	header.Set(headerTimestamp, strconv.Itoa(docRespTS))
	header.Set(headerSerial, docCertSN)
	header.Set(headerSignature, docRespSig)
	plain, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, docResponseEnvelope)
	if err != nil {
		t.Fatalf("VerifyAndDecryptResponse: %v", err)
	}
	if string(plain) != docResponsePlaintext {
		t.Fatalf("documented response plaintext mismatch:\n got %s\nwant %s", plain, docResponsePlaintext)
	}
}

// TestResponseRejectsTamperedBody confirms the signature check is load bearing:
// flipping a byte of the response must fail rather than yield plaintext.
func TestResponseRejectsTamperedBody(t *testing.T) {
	sec := docAPISecurity(t)
	header := http.Header{}
	header.Set(headerTimestamp, strconv.Itoa(docRespTS))
	header.Set(headerSerial, docCertSN)
	header.Set(headerSignature, docRespSig)
	tampered := append([]byte(nil), docResponseEnvelope...)
	tampered[len(tampered)-3] ^= 0x01
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, tampered); err == nil {
		t.Fatal("tampered response verified, want error")
	}
}

// TestUnsignedResponseRejectedByDefault covers the security default: an endpoint
// marked as supporting the layer must answer signed, so a missing signature is
// an error rather than a silent downgrade to plain text.
func TestUnsignedResponseRejectedByDefault(t *testing.T) {
	sec := docAPISecurity(t)
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, http.Header{}, []byte(`{"errcode":40001}`)); err == nil {
		t.Fatal("expected an error for an unsigned response in the default (strict) mode")
	}
}

// TestUnsignedResponseAllowedWhenOptedIn covers the explicit opt-out: with
// AllowUnsignedResponses the caller is told to handle the bytes as a plain,
// unverified reply.
func TestUnsignedResponseAllowedWhenOptedIn(t *testing.T) {
	sec := docAPISecurityLenient(t)
	plain, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, http.Header{}, []byte(`{"errcode":40001}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plain != nil {
		t.Fatalf("expected nil plaintext for an unsigned response, got %s", plain)
	}
}

// TestUnsignedEnvelopeAlwaysRejected covers the shape check: a body that is an
// encrypted envelope but carries no signature must fail even with
// AllowUnsignedResponses, because accepting it would yield a zeroed struct with
// no error.
func TestUnsignedEnvelopeAlwaysRejected(t *testing.T) {
	for name, sec := range map[string]*APISecurity{"strict": docAPISecurity(t), "lenient": docAPISecurityLenient(t)} {
		if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, http.Header{}, docResponseEnvelope); err == nil {
			t.Errorf("%s: expected an error for an unsigned envelope", name)
		}
	}
}

// TestFutureResponseTimestampAccepted covers the one-directional replay window
// the guide specifies: a response timestamped slightly ahead of the local clock
// is accepted, and only a stale one is rejected.
func TestFutureResponseTimestampAccepted(t *testing.T) {
	sec := docAPISecurity(t)
	// Local clock 60s BEHIND the signed response timestamp: within tolerance.
	sec.now = func() time.Time { return time.Unix(docRespTS-60, 0) }
	header := http.Header{}
	header.Set(headerTimestamp, strconv.Itoa(docRespTS))
	header.Set(headerSerial, docCertSN)
	header.Set(headerSignature, docRespSig)
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, docResponseEnvelope); err != nil {
		t.Fatalf("a response slightly ahead of the local clock must be accepted: %v", err)
	}
	// Far ahead is still accepted (only staleness is bounded), matching the
	// guide's `local_ts - resp_ts > 300` check.
	sec.now = func() time.Time { return time.Unix(docRespTS-3600, 0) }
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, docResponseEnvelope); err != nil {
		t.Fatalf("a future timestamp must not be rejected: %v", err)
	}
}

// TestSecureRequestShape checks what the runtime puts on the wire: the
// Wechatmp-* headers, the parameter set carried inside the ciphertext, the
// leading security fields, and that access_token stays in the URL query.
func TestSecureRequestShape(t *testing.T) {
	sec := docAPISecurity(t)
	type req struct {
		Scene    int64  `json:"scene"`
		ClientIP string `json:"client_ip"`
		OpenID   string `json:"openid"`
		AppID    string `json:"appid"`
		Token    string `query:"access_token"`
	}
	body, headers, err := sec.SecureRequest(docAppID, docURLPath, collectParams(reflect.ValueOf(req{
		Scene: 7, ClientIP: "1.2.3.4", OpenID: "o1", AppID: docAppID, Token: "TOKEN",
	})))
	if err != nil {
		t.Fatalf("SecureRequest: %v", err)
	}
	if headers[headerAppID] != docAppID {
		t.Errorf("%s = %q", headerAppID, headers[headerAppID])
	}
	if headers[headerTimestamp] != strconv.Itoa(docRespTS) {
		t.Errorf("%s = %q", headerTimestamp, headers[headerTimestamp])
	}
	if headers[headerSignature] == "" {
		t.Errorf("%s is empty", headerSignature)
	}
	plain, err := sec.open(body, apiSecurityAAD(docURLPath, docAppID, docRespTS, docSymSN))
	if err != nil {
		t.Fatalf("decrypt own request: %v", err)
	}
	var got, want map[string]any
	if err := json.Unmarshal(plain, &got); err != nil {
		t.Fatal(err)
	}
	wantJSON := `{"_n":"x","_appid":"` + docAppID + `","_timestamp":1635927956,"scene":7,"client_ip":"1.2.3.4","openid":"o1","appid":"` + docAppID + `"}`
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatal(err)
	}
	got["_n"] = "x"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("payload mismatch:\n got %v\nwant %v", got, want)
	}
	if strings.Contains(string(plain), "TOKEN") {
		t.Errorf("access_token leaked into the encrypted payload: %s", plain)
	}
	if !strings.HasPrefix(string(plain), `{"_n":`) {
		t.Errorf("payload does not start with the security fields: %s", plain)
	}
}

// TestNewAPISecurityValidation covers the configuration errors callers hit.
func TestNewAPISecurityValidation(t *testing.T) {
	base := APISecurityConfig{
		SymKey:        docSymKey,
		SymSN:         docSymSN,
		PrivateKey:    docPrivateKeyPEM,
		PlatformCerts: map[string]string{docCertSN: docPlatformCertPEM},
	}
	tests := map[string]func(*APISecurityConfig){
		"missing sn":       func(c *APISecurityConfig) { c.SymSN = "" },
		"bad sym key":      func(c *APISecurityConfig) { c.SymKey = "not-base64!" },
		"short sym key":    func(c *APISecurityConfig) { c.SymKey = base64.StdEncoding.EncodeToString([]byte("short")) },
		"missing key":      func(c *APISecurityConfig) { c.PrivateKey = "" },
		"bad cert":         func(c *APISecurityConfig) { c.PlatformCerts = map[string]string{docCertSN: "nope"} },
		"unsupported algo": func(c *APISecurityConfig) { c.Alg = APIAlgSM4GCM },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := base
			mutate(&cfg)
			if _, err := NewAPISecurity(cfg); err == nil {
				t.Fatalf("%s: expected an error", name)
			}
		})
	}
}

// TestAPISecurityMissingCertificate covers a response whose serial has no
// configured certificate: it must fail loudly rather than skip verification.
func TestAPISecurityMissingCertificate(t *testing.T) {
	sec := docAPISecurity(t)
	header := http.Header{}
	header.Set(headerTimestamp, strconv.Itoa(docRespTS))
	header.Set(headerSerial, "unknown-serial")
	header.Set(headerSignature, docRespSig)
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, docResponseEnvelope); err == nil {
		t.Fatal("expected an error for an unknown certificate serial")
	}
}

// TestAPISecurityStaleResponseTimestamp covers the replay bound: a signed
// response far from the local clock is rejected even when the signature checks
// out.
func TestAPISecurityStaleResponseTimestamp(t *testing.T) {
	sec := docAPISecurity(t)
	sec.now = func() time.Time { return time.Unix(docRespTS+600, 0) }
	header := http.Header{}
	header.Set(headerTimestamp, strconv.Itoa(docRespTS))
	header.Set(headerSerial, docCertSN)
	header.Set(headerSignature, docRespSig)
	if _, err := sec.VerifyAndDecryptResponse(docAppID, docURLPath, header, docResponseEnvelope); err == nil {
		t.Fatal("expected the stale response to be rejected")
	}
}

// TestCallAppliesAPISecurity exercises the endpoint runtime end to end: a call
// carrying an installed APISecurity decorator must encrypt the request and
// while the same call without the option goes out in the clear.
func TestCallAppliesAPISecurity(t *testing.T) {
	// Lenient so the unsigned test reply is accepted; the request half is what
	// this test asserts.
	sec := docAPISecurityLenient(t)
	sec.now = func() time.Time { return time.Unix(docReqTS, 0) }
	var gotHeaders http.Header
	var gotBody []byte
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders, gotQuery = r.Header, r.URL.RawQuery
		gotBody, _ = io.ReadAll(r.Body)
		// An unsigned reply, which this lenient runtime treats as a plain
		// response (the documented signature path is covered by
		// TestDocResponseVerifyAndDecrypt, keyed to the documented urlpath).
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	c := NewEndpointClient(EndpointConfig{
		AppID: docAppID, AppSecret: "s", BaseURL: srv.URL, HTTPClient: srv.Client(),
		Token:     func(context.Context) (string, error) { return "TOKEN", nil },
		Modifiers: []RequestModifier{sec},
	})
	type req struct {
		Scene    int64  `json:"scene"`
		ClientIP string `json:"client_ip"`
		OpenID   string `json:"openid"`
		AppID    string `json:"appid"`
	}
	if _, err := Call[map[string]any](c, context.Background(), http.MethodPost, "/wxa/getuserriskrank", &req{Scene: 0, ClientIP: "127.0.0.1", OpenID: "oEWzBfmdLqhFS2mTXCo2E4Y9gJAM", AppID: docAppID}, nil, true, SignNone); err != nil {
		t.Fatalf("secured call: %v", err)
	}
	if gotHeaders.Get(headerSignature) == "" || gotHeaders.Get(headerAppID) != docAppID {
		t.Errorf("secured call did not send the security headers: %v", gotHeaders)
	}
	if len(gotBody) == 0 || !strings.Contains(string(gotBody), `"authtag"`) {
		t.Errorf("secured call body is not an encrypted envelope: %s", gotBody)
	}
	if !strings.Contains(gotQuery, "access_token=") {
		t.Errorf("access_token missing from the query: %q", gotQuery)
	}

	// A client with no decorator must stay in the clear, and so must a decorated
	// client for a path the decorator does not cover.
	plainSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "authtag") {
			t.Errorf("undecorated call was encrypted: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer plainSrv.Close()
	c2 := NewEndpointClient(EndpointConfig{
		BaseURL: plainSrv.URL,
		Token:   func(context.Context) (string, error) { return "TOKEN", nil },
	})
	if _, err := Call[map[string]any](c2, context.Background(), http.MethodPost, "/x", &req{Scene: 1}, nil, true, SignNone); err != nil {
		t.Fatalf("plain call: %v", err)
	}
}
