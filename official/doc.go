// Package official is a typed Go client for the WeChat Official Account
// (公众号 / 服务号) server APIs.
//
// Reference docs:
//   - 公众号: https://developers.weixin.qq.com/doc/subscription/api/
//   - 服务号: https://developers.weixin.qq.com/doc/service/guide/
//
// Like the miniprogram package, it is built on the shared
// github.com/go-sphere/weixin-api/core package: the OfficialAccount struct
// embeds *core.Client so access-token management, the universal
// {"errcode":..,"errmsg":..} envelope handling and every HTTP helper are
// promoted onto it.
//
// # Conventions
//
//   - Methods that require an access token acquire (and cache) it automatically
//     through core, and transparently retry once after a token-expiry error.
//   - Every method documents the exact upstream reference page with a
//     "Reference:" URL.
package official
