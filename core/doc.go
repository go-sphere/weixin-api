// Package core provides the shared client foundation used by every WeChat
// platform package in this module (miniprogram for Mini Programs, official for
// 公众号 / 服务号 official accounts, and future platforms).
//
// It implements the pieces that are identical across all WeChat server APIs:
//
//   - the plain net/http transport with JSON/form/raw/upload support;
//   - the universal {"errcode":..,"errmsg":..} response envelope and its
//     error classification (sentinel errors for token problems, *APIError
//     otherwise);
//   - access-token and JS-SDK ticket management with cache + singleflight
//     deduplication and transparent retry after token-expiry errors;
//   - a pluggable Cache interface (an in-process memory cache is included).
//
// Platform packages embed *Client so their endpoint methods can call the
// promoted Do/GetJSON/PostJSON/WithToken* helpers directly.
package core
