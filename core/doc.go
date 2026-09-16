// Package core is the shared runtime and foundation of the WeChat API clients
// in this module. The platform packages (miniprogram for Mini Programs,
// official for 公众号 / 服务号 official accounts) are generated code that
// delegates every request to this package, so all cross-platform behaviour
// lives here exactly once.
//
// It provides:
//
//   - the plain net/http transport (JSON bodies, form fields, raw bytes and
//     multipart uploads) with proxy and timeout configuration;
//   - access-token management: cache with singleflight de-duplication and a
//     transparent single retry after a token-expiry error;
//   - the universal {"errcode":..,"errmsg":..} envelope and its
//     classification — token problems collapse into sentinel errors that
//     errors.Is matches, every other code becomes an *APIError carrying the
//     code, message, rid and raw body;
//   - EndpointClient, the runtime behind the generated endpoint methods:
//     tag-driven request assembly, credential and token injection, request
//     signing (XPay pay_sig and user-level signatures), and documented
//     error-code enrichment;
//   - message-callback crypto (signature verification and AES-CBC
//     (de)encryption) and JS-SDK config signing, neither of which can be
//     derived from the HTTP documentation;
//   - a pluggable Cache interface with an in-process memory implementation.
//
// Callers normally use a platform package rather than this one directly; core
// is imported for the shared types those packages expose (APIError, Cache,
// Credentials, MessageCrypto, MiniAppEnv, ...).
package core
