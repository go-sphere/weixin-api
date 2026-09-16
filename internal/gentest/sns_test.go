package gentest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sphere/weixin-api/official"
)

// sns/userinfo carries the USER's OAuth token. The generated client must send
// the caller's value and must not fetch or inject the application token.
func TestSnsUserinfoUsesCallerToken(t *testing.T) {
	var query string
	var tokenFetched bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/cgi-bin/token") {
			tokenFetched = true
			_, _ = io.WriteString(w, `{"access_token":"APPPTOKEN","expires_in":7200}`)
			return
		}
		query = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"openid":"o","nickname":"n"}`)
	}))
	t.Cleanup(srv.Close)
	c := official.New(official.Config{AppID: "a", AppSecret: "b", BaseURL: srv.URL})
	if _, err := c.GetSnsUserinfo(t.Context(), &official.GetSnsUserinfoRequest{
		AccessToken: "USERTOKEN",
		OpenID:      "o",
	}); err != nil {
		t.Fatal(err)
	}
	if tokenFetched {
		t.Error("app token was fetched for an endpoint that needs the user token")
	}
	if !strings.Contains(query, "access_token=USERTOKEN") {
		t.Errorf("user token not sent: %s", query)
	}
}
