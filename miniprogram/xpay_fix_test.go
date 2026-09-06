package miniprogram

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// TestXPayRequestContentTypeAndSignatures asserts two invariants of the XPay
// transport (RF-022):
//  1. the pre-signed body is sent with the JSON content type (not octet-stream);
//  2. pay_sig/signature are deterministic for a fixed body, so a golden vector
//     can pin the algorithm and prevent future drift.
func TestXPayRequestContentTypeAndSignatures(t *testing.T) {
	var (
		contentType string
		rawQuery    string
		sentBody    string
	)
	w, _ := recordClient(t, func(rw http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		rawQuery = r.URL.RawQuery
		buf := make([]byte, 0, 512)
		tmp := make([]byte, 256)
		for {
			n, err := r.Body.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if err != nil {
				break
			}
		}
		sentBody = string(buf)
		writeJSON(rw, map[string]any{"errcode": 0, "errmsg": "ok", "balance": 10})
	})
	// Reconfigure the recorded client with an AppKey (recordClient uses none).
	w.config.AppKey = "test-app-key"

	req := &XPayQueryUserBalanceRequest{XPayCommon: XPayCommon{OpenID: "o1", Env: XPayEnvProduction}}
	resp, err := w.XPayQueryUserBalance(t.Context(), req, "session-key")
	if err != nil {
		t.Fatalf("XPayQueryUserBalance: %v", err)
	}
	if resp.Balance != 10 {
		t.Fatalf("unexpected response %+v", resp)
	}

	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	q := parseQuery(t, rawQuery)
	paySig := q["pay_sig"]
	sig := q["signature"]
	if paySig == "" || sig == "" {
		t.Fatalf("missing signatures in query %s", rawQuery)
	}

	// The body must be exactly the bytes that were signed. Recompute the
	// signature the same way xpayRequest does and compare.
	var expectedBody map[string]any
	if err := json.Unmarshal([]byte(sentBody), &expectedBody); err != nil {
		t.Fatalf("sent body is not JSON: %v (%q)", err, sentBody)
	}
	wantPaySig := hmacSHA256Hex([]byte("test-app-key"), []byte("/xpay/query_user_balance&"+sentBody))
	if paySig != wantPaySig {
		t.Fatalf("pay_sig mismatch:\n got %q\nwant %q (body %q)", paySig, wantPaySig, sentBody)
	}
	wantUserSig := hmacSHA256Hex([]byte("session-key"), []byte(sentBody))
	if sig != wantUserSig {
		t.Fatalf("user signature mismatch:\n got %q\nwant %q", sig, wantUserSig)
	}
}

func parseQuery(t *testing.T, raw string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, part := range strings.Split(raw, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}
	return out
}
