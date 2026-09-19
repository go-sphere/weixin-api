# weixin-api

A typed, idiomatic Go client suite for the WeChat **server APIs**. Each WeChat
platform gets its own package built on a shared HTTP/token foundation. The
platform packages are **generated** from the official documentation by
[`tools/docsync`](./tools/docsync) — do not edit them by hand; run
`make docs-sync` and the change appears (or upstream changed).

| Package | Platform | Reference |
| --- | --- | --- |
| [`miniprogram`](./miniprogram) | 微信小程序 Mini Program | https://developers.weixin.qq.com/miniprogram/dev/server/API/ |
| [`official`](./official) | 公众号 / 服务号 Official Account | https://developers.weixin.qq.com/doc/subscription/api/ · https://developers.weixin.qq.com/doc/service/guide/ |
| [`core`](./core) | shared foundation (transport, tokens, errors, crypto, JS-SDK) | — |
| [`internal/gentest`](./internal/gentest) | package-level invariants for the generated code | — |
| [`tools/docsync`](./tools/docsync) | documentation crawler, extractor and code generator | — |

The clients use a plain `net/http` transport — **no third party HTTP library** —
and every endpoint method carries typed request/response structs plus a `Doc:`
link straight to its official WeChat documentation page. Regeneration is
byte-for-byte reproducible, so `git diff` over the generated packages shows
exactly what changed upstream.

Regenerate with `make docs-sync`; check for upstream drift with
`make docs-check`; run the pipeline and wire-level test suites with
`make docs-test` and `make docs-parity`.

The pipeline runs in stages, so the slow network crawl happens once into a
gitignored local cache and the generator can then be iterated offline:

```sh
make docs-crawl    # crawl the doc site -> .docs-cache/ + manifest.json (network)
make docs-gen      # rebuild miniprogram/ and official/ from the cache (offline)
make docs-swagger  # rebuild the OpenAPI contracts from the cache (offline)
IGNORE_CACHE=1 make docs-gen   # force a re-crawl first
```

### Generated client configuration

`miniprogram.New(Config{...})` / `official.New(Config{...})` accept:

- `AppID` / `AppSecret` — the client fetches, caches and refreshes the
  `access_token` itself; request structs do not expose that parameter, so it
  cannot be blanked out by a zero value.
- `Token` — supply your own access-token source instead (no caching/refresh).
- `AppKey` — the 米大师 virtual-payment signing key. Endpoints under `/xpay/`
  are signed automatically: `pay_sig = HMAC-SHA256(AppKey, path + "&" + body)`,
  plus `signature = HMAC-SHA256(SessionKey, body)` for user-level calls (set
  the generated `SessionKey` field on the request). A signed call fails fast
  when `AppKey` is missing rather than sending an unsigned request.
- `Env` — the Mini Program environment (`release` / `trial` / `develop`); the
  subscribe-message API defaults its `miniprogram_state` from it.
- `Modifiers` — optional `[]core.RequestModifier` decorators. A modifier sees
  the call before it is serialized (typed request values, path, method) and
  after the response arrives, so optional layers the endpoint documentation
  does not describe can be installed without the generated code knowing about
  them. `core.APISecurity` is one such decorator (see "API security" below);
  `Client.Use` installs more on an existing client.
- `Cache`, `Proxy`, `BaseURL`, `HTTPClient` — optional overrides.

WeChat Pay `retail/B2b` endpoints instead take a caller-computed `PaySig`
field, because the merchant's payment key is not the app key.

#### API security (二次加密和签名)

WeChat lets an account turn on API 加密 in the MP backend. The endpoints whose
documentation carries the "支持加密请求" note then require their parameters to
be encrypted into the body and the body to be signed; the response arrives
encrypted and signed too. `core.APISecurity` implements this as a
`RequestModifier` decorator, so the generated code stays untouched. Because
WeChat documents support per API ("只有部分 API 支持加解密，具体可参考各 API
文档"), the caller names the paths its account actually uses:

```go
sec, err := core.NewAPISecurity(core.APISecurityConfig{
	SymKey:     "<base64 API 对称密钥>",
	SymSN:      "<API 对称密钥编号>",
	PrivateKey: "<PEM API 应用私钥>",
	// 平台证书编号 -> PEM, downloaded from the MP backend.
	PlatformCerts: map[string]string{"<编号>": "<PEM 平台证书>"},
	// The endpoints to secure. Fill this with the paths your account uses
	// whose docs carry the "支持加密请求" note.
	Paths: []string{"/wxa/getuserriskrank", "/wxa/msg_sec_check", /* ... */},
})
if err != nil {
	return err
}
client := miniprogram.New(miniprogram.Config{
	AppID: ..., AppSecret: ...,
	Modifiers: []core.RequestModifier{sec},
})
```

Any call whose path is not in `Paths` is left exactly as the runtime assembled
it, so the decorator is scoped to what the caller declares. `Client.Use(sec)`
installs one after construction.

The wire protocol (`urlpath\nappid\ntimestamp\nbody` for RSA-PSS signing,
`urlpath|appid|timestamp|sn` as GCM additional data, the `_n`/`_appid`/
`_timestamp` security fields) is pinned in `core/apisecurity_doc_test.go`
against the vectors the guide publishes. Parameters come from the typed request
struct, so a field's declared type is preserved in the encrypted JSON;
`access_token` is kept out of the payload and in the URL query, as the guide
requires.

By default a covered endpoint must answer with a signed response: if the reply
carries no `Wechatmp-Signature` the call fails rather than silently falling back
to unverified plain text, which is what "防篡改" is for. A body shaped like an
encrypted envelope is never accepted unsigned, because that would surface as a
successful call with zeroed fields. If a deployment genuinely receives unsigned
replies, set `APISecurityConfig.AllowUnsignedResponses: true` to opt out.

Four endpoints that the docs mark as supporting the layer are GET methods
(`/wxa/business/getuserencryptkey` and the three
`/cgi-bin/express/business/*/getall`): the guide only demonstrates a POST, so
the GET-with-encrypted-body path is an inference and has not been verified
against the live platform. If WeChat rejects it, simply leave those paths out of
`Paths` to call them in the clear as before.

Only AES256_GCM + RSAwithSHA256 is implemented. The 国密 pair (SM4_GCM +
SM2withSM3) is reserved but not implemented: the guide publishes no verifiable
vector for it and Go has no standard-library implementation, so the signature
encoding could not be confirmed against WeChat.

## Packages

### core

The shared runtime behind every platform package. The generated clients
delegate all cross-platform behaviour here, so it exists once rather than once
per platform:

- `core.EndpointClient`: the runtime of the generated endpoint methods —
  tag-driven request assembly (query vs JSON body), application-credential and
  access-token injection, and request signing.
- `core.RequestModifier` + `Client.Use`: the decorator hook that sees a call
  before serialization (typed request values) and after the response, letting
  optional layers attach without generated-code changes.
- `core.APISecurity` + `core.NewAPISecurity()`: the API 二次加密和签名 layer as
  one such decorator — AES-256-GCM request encryption, RSA-PSS request signing,
  and platform-certificate response verification and decryption, scoped to the
  paths the caller lists.
- `core.Client`: the authenticated HTTP client beneath it; plain `net/http`
  transport, proxy/timeout configuration, and per-account `access_token`
  caching with singleflight de-duplication plus a transparent retry after a
  token-expiry error.
- `core.APIError` / `core.ErrResponse`: the universal
  `{"errcode":..,"errmsg":..}` envelope. Token failures collapse into sentinel
  errors (`core.ErrorInvalidCredential`, ...) for `errors.Is`; other codes keep
  the code, message, `rid` and raw body.
- `core.Cache` + `core.NewMemoryCache()`: pluggable credential cache.
- `core.MessageCrypto`: message-push callback signature verification and
  AES-256-CBC payload (de)encryption, including the signed reply envelope
  (`core.Reply`). Also `core.JSSDKConfig` / `core.JSSDKSignature` for JS-SDK
  signing.
- `core.MiniAppEnv`: deployment environment used by the subscribe-message APIs.

### miniprogram

The Mini Program server API client (moved from the former single `wechat`
package, now one method per upstream endpoint across the whole catalogue:
login/session, subscribe messages, QR codes & URL links, content security,
customer service, datacube, express/intracity logistics, order shipping &
transaction guarantee, live broadcast, plugin & nearby POI, hardware/IoT,
XPay virtual payment, CloudBase, novel/short-drama, city service and more).

```go
client := miniprogram.NewMiniProgram(miniprogram.Config{
	AppID:     "wx1234567890abcdef",
	AppSecret: "0123...",
}, nil) // nil cache -> in-process memory cache

session, err := client.JsCode2Session(ctx, "wx-login-code")
```

### official

Official Account (公众号/服务号) APIs: user & tag management, blacklist,
customer-service messages, template messages, mass messages, media & permanent
materials, QR codes & short URLs, custom menus, draft box & publishing,
article/user/interface analysis, callback IP/quota and JS-SDK signing.

```go
oa := official.NewOfficialAccount(official.Config{
	AppID:     "wxb0000000000000",
	AppSecret: "abcd...",
}, nil)

info, err := oa.GetUserInfo(ctx, &official.GetUserInfoRequest{OpenID: "o1"})
```

## Conventions

- Methods that require an access token acquire (and cache) it automatically and
  transparently retry once after a token-expiry style error.
- JSON responses are checked for the universal error envelope; a non-zero code
  is surfaced as an error instead of silently returning partial data.
- Binary endpoints (QR codes, media files) return raw bytes and automatically
  surface an embedded JSON business error.
- Every exported method documents its parameters, response shape and a
  `Reference:` URL to the official page.

## Development

```sh
make fmt      # gofmt + goimports
make lint     # go vet + golangci-lint + nilaway
make test     # unit tests (httptest based, no network access)
make check    # everything CI runs
```

## License

MIT
