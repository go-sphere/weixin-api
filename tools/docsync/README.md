# docsync

A maintenance pipeline that mirrors the Chinese WeChat server-API
documentation, extracts every documented endpoint into per-platform
**OpenAPI 3.0 contracts**, and generates a **Go client package per platform**
from those contracts:

```
HTML pages ──crawl──▶ .docs-cache/ + manifest.json ──▶ deterministic extractor ──▶ []Operation
 (870 pages)   (network)     (gitignored / committed)      (unified doc template)        │
                                     └── LLM fallback (-llm) ──────────────────────────┘
                                                                                        │
                                       swagger/miniprogram.swagger.{json,yaml} ◀── swagger stage
                                       swagger/official.swagger.{json,yaml}
                                                                                        │
                                          miniprogram/*.go, official/*.go      ◀── gen stage
```

| Tree | Seed | Generated package |
| --- | --- | --- |
| Mini Program server API | https://developers.weixin.qq.com/miniprogram/dev/server/API/ | `miniprogram` (430 ops) |
| Service Account API | https://developers.weixin.qq.com/doc/service/api/ | `official` (196 ops, merged) |
| Subscription Account API | https://developers.weixin.qq.com/doc/subscription/api/ | `official` (merged) |

The generated packages live **inside the library module** (`miniprogram/`,
`official/`), not in the tool module, so consumers can import them directly.
They depend on `core`, the shared runtime.

Note: the Mini Program tree lives under `/miniprogram/dev/server/API/`; the
old `/miniprogram/dev/api-backend/` URL is a legacy alias with broken links.

## Stages

The pipeline is split so the network crawl — the only slow, non-reproducible
step — runs **once** into a local cache, and the extractor and the generator
are then iterated offline against that cache:

| Stage | Command | Input | Output | Network |
| --- | --- | --- | --- | --- |
| crawl | `make docs-crawl` | the live doc site | `.docs-cache/`, `manifest.json` | yes |
| swagger | `make docs-swagger` | `.docs-cache/` | `swagger/*.{json,yaml}` | no |
| gen | `make docs-gen` | `.docs-cache/` | `miniprogram/`, `official/` | no |
| all | `make docs-sync` | the live doc site | all of the above | yes |

The page cache (`.docs-cache/`, gitignored, ~130 MB, safe to delete) stores the
**raw response bytes**, one file per page, mirroring the URL path. That is
exactly the input `parsePage` consumed during the crawl, so an offline stage is
byte-for-byte equivalent to a live run.

The offline stages validate the cache against the committed manifest: a page
that is missing fails with a pointer to `make docs-crawl`, and one whose
content no longer matches its recorded hash is reported as a warning (the
artifacts are still rebuilt, from the cached markup).

Set `IGNORE_CACHE=1` to force a re-crawl before an offline stage:

```sh
IGNORE_CACHE=1 make docs-gen   # re-crawl, then regenerate the Go clients
```

## Pipeline stages

1. **Crawl + change detection** — every page's document body (TOC excluded)
   is normalized and hashed into `tools/docsync/manifest.json` (committed);
   the raw bodies are cached under `.docs-cache/` (gitignored). A `git diff`
   of the manifest after a sync shows exactly which pages changed upstream.
2. **Deterministic extraction** — all three doc trees share one page
   template (`1.调用方式 / 2.请求参数 / 3.返回参数 / 6.错误码`), so a fixed
   parser recovers `method`, `path`, query/body/response fields (names,
   normalized types, required flags, Chinese descriptions) and error codes.
   Measured coverage: 775 of 776 leaf pages; the one outlier is the LLM
   fallback's job.
   Nested members documented in separate `Res.* Object Payload` /
   `Body.* Object Payload` blocks are reattached to their parent field
   (376 of 870 pages, up to 5 levels deep, including arrays of objects), so
   the contract and the generated structs are nested rather than flattened
   into `map[string]any`.
3. **LLM fallback (optional, `-llm`)** — pages the fixed parser cannot read
   are sent to an OpenAI-compatible endpoint (DeepSeek; credentials in the
   gitignored `LLM.ENV`) via `openai-go` with a strict tool-call JSON schema.
   **The model only fills in endpoint facts.** Operation IDs and every
   generated identifier derive from the endpoint path by fixed rules
   (`assignOperationIDs`), so LLM randomness can never change the API
   surface. Default: off.
4. **OpenAPI emission** — one OpenAPI 3.0 document per platform
   (`swagger/miniprogram.swagger.{json,yaml}`,
   `swagger/official.swagger.{json,yaml}`; committed,
   `linguist-generated=true`), built with `kin-openapi` and validated against
   the spec before writing. The service and subscription trees merge into the
   official contract; an endpoint documented in both trees appears once.
5. **Go generation** — one runnable client package per platform
   (`miniprogram/`, `official/`; committed,
   `linguist-generated=true`): one method + request/response structs per
   operation, with the Chinese field descriptions as doc comments. Emitted
   with `dave/jennifer` (automatic import management and formatting) and
   verified with `go vet` as part of the run. Separate packages mean separate
   `Client` instances — a mini program appid and an official account appid
   never share a token.
6. **Binary-response endpoints** — pages whose notes say success returns raw
   bytes ("会直接返回图片二进制内容", e.g. the mini program code APIs) are
   detected during extraction: the OpenAPI contract declares a binary 200
   response plus a default error envelope, and the generated method returns
   `([]byte, error)` through `callBinary`, which surfaces only failures as
   `*APIError`.
7. **Multipart uploads** — request fields documented as `formdata` become an
   `Upload` file part: the contract declares `multipart/form-data` and the
   generated method takes `*Upload` (29 endpoints).
8. **Typed scalars and naming** — `number` maps to `int64` unless the
   description implies a fraction (lat/long, ratios) in which case it stays
   `float64`; fields keep the docs' own casing (`phoneNumber` →
   `PhoneNumber`, `appid` → `AppID`). Documented required fields carry no
   `omitempty`, so a required zero value is still serialized.
9. **Generated runtime** — the client caches the access token, de-duplicates
   concurrent fetches (singleflight) and retries once with a refreshed token
   when WeChat reports 40001/40014/42001; non-GET calls always carry the JSON
   content type, and media downloads return raw bytes.
10. **Credential awareness** — the docs declare `access_token` either as a
    query parameter or inline in the call URL; endpoints that authenticate
    with appid+secret (`/sns/jscode2session`, `/sns/oauth2/...`) or none at
    all (`/cgi-bin/token`) are detected and never sent a token. When a request
    struct declares `appid`/`secret`, the client fills them from its `Config`
    the way the hand-written clients do.
11. **Reviewed type overrides** — a small fixed table
    (`typeOverrides` in extract.go) corrects field types upstream states
    wrongly, e.g. `ip_list` documented as `string` but returned as an array,
    or `tagid_list` documented as a bare array but returning numeric ids. The
    corrections recurse into nested members. Entries are hand-written and cite
    the evidence, so output stays reproducible.
12. **Request signing** — `/xpay/` endpoints are signed by the runtime:
    `pay_sig = HMAC-SHA256(Config.AppKey, path + "&" + rawBody)`, plus
    `signature = HMAC-SHA256(session_key, rawBody)` for user-level calls (the
    generated request carries a `SessionKey` field for it). `access_token` and
    `pay_sig` are therefore client-owned and are not emitted as request fields,
    so a zero value can never overwrite the credential the runtime produced.
    `retail/B2b` endpoints instead take a caller-computed `PaySig`, because the
    merchant payment key differs from the app key.
13. **Error-code documentation** — each page's "6.错误码" table becomes an
    `x-error-codes` extension on the OpenAPI operation and a package-level
    `map[int]ErrDoc` in the generated package. `call`/`callBinary` look the
    errcode up on failure and fill `APIError.Description`, `APIError.Solution`
    and `APIError.Doc`, so the documented meaning of a failure is one field
    away.

### Naming stability

Generated identifiers are pure functions of the endpoint, e.g.
`GET /cgi-bin/stable_token` → `GetCgiBinStableToken` (split on `-`, `_`, `/`,
`.` and letter/digit boundaries, CamelCase, HTTP verb prefix; collisions get
a numeric suffix in deterministic order). The same set of endpoints always
produces byte-identical output — verified by running the pipeline twice.
IDs are assigned per package: two platforms may reuse a name when both
document similarly-shaped endpoints.

## Usage

```sh
make docs-crawl   # crawl only: refresh .docs-cache/ + manifest.json (network)
make docs-swagger # contracts only, from the local cache (offline)
make docs-gen     # Go clients only, from the local cache (offline)
make docs-sync    # all three stages, in that order
make docs-check   # crawl + report drift only; exit 1 when upstream changed
make docs-test    # unit tests for the pipeline itself
make docs-parity  # generated vs hand-written request parity suite
```

### Parity suite

`tools/docsync/parity` points both the generated and the hand-written clients
at one httptest server, feeds them identical inputs and diffs the request each
one actually sends (method, path, query, body, content type). It covers
callback-check, callback-IP, media download, multipart upload and the
appid+secret login flow. It is the safety net for migrating a hand-written
package onto its generated counterpart: path, parameter naming, auth handling
or encoding drift fails the suite first.

The suite has two halves:

- `coverage_test.go` statically audits the whole surface — every endpoint the
  hand-written packages call must exist in the generated package with the same
  verb (427/430 for miniprogram, 56/59 for official). The residue lives in a
  reviewed `knownGaps` table where each entry states why (legacy API no longer
  documented, superseded endpoint, or a path documented in the other tree); a
  *new* gap fails the audit.
- `parity_test.go` compares requests on the wire for shape-family
  representatives: callback check, callback IP, media download, multipart
  upload, appid+secret login, query-only GET, simple JSON body, nested JSON
  body.
- `TestSharedCryptoAndJSSDK` pins the two capabilities that cannot be derived
  from the HTTP documentation — message-callback crypto and the JS-SDK
  signature — to the shared `core` runtime: a ciphertext produced by the
  hand-written crypto must decrypt with the generated one, and both packages
  expose the same signing algorithm.

Recorded divergences (deliberate, not silent): empty-POST body encoding
(hand-written sends `{}`, generated omits the body; WeChat accepts both) and
two response shapes the docs describe incompletely (`ip_list` typed as bare
`array`; `ip_list` on `getcallbackip` typed `string` but returned as an
array — the latter is corrected by the `typeOverrides` table).

Directly: `cd tools/docsync && go run . [-llm] [-check] [-format=json]`.

- A tracked page that fails to fetch aborts the run (a partial network
  failure must never look like "page removed"); broken links on doc pages are
  reported as skips.
- `-llm` needs `LLM.ENV` at the repository root
  (`OPENAI_API_BASE`/`OPENAI_API_KEY`, optional `OPENAI_MODEL`).

## Typical workflow

1. `make docs-check` — no drift means the clients match the current docs.
2. On drift, `git diff tools/docsync/manifest.json` lists the pages;
   `git diff` over `swagger/`, `miniprogram/` and `official/` shows the exact
   contract and code consequences (new operations, changed fields, new error
   codes).
3. Port the changes into the hand-written packages (`miniprogram/`,
   `official/`) or adopt the generated clients directly.
