package miniprogram

import (
	"io"

	"github.com/go-sphere/weixin-api/core"
)

// requestOptions controls per-call behaviour of access-token based methods.
// It is the miniprogram-side view of core.RequestOptions.
type requestOptions struct {
	retryable         bool
	reloadAccessToken bool
}

// defaultReqOptions returns the default options: retries enabled and no forced
// token reload.
func defaultReqOptions() requestOptions {
	return requestOptions{retryable: true}
}

// toCore converts the package-private options into core.RequestOptions.
func (o requestOptions) toCore() core.RequestOptions {
	return core.RequestOptions{
		Retryable:         o.retryable,
		ReloadAccessToken: o.reloadAccessToken,
	}
}

// decodeJSON is a thin alias of core.DecodeJSON kept so endpoint files stay
// short.
func decodeJSON(data []byte, dst any) error {
	return core.DecodeJSON(data, dst)
}

// jsonMarshal marshals v to JSON bytes.
func jsonMarshal(v any) ([]byte, error) {
	return core.JSONMarshal(v)
}

// hmacSHA256Hex computes the lowercase hex HMAC-SHA256 of msg keyed by key.
func hmacSHA256Hex(key, msg []byte) string {
	return core.HMACSHA256Hex(key, msg)
}

// readerOf adapts any value satisfying the minimal reader shape to io.Reader.
func readerOf(r interface{ Read([]byte) (int, error) }) io.Reader {
	if rr, ok := r.(io.Reader); ok {
		return rr
	}
	return nil
}
