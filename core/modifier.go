package core

import (
	"errors"
	"net/http"
)

// This file holds the request/response decoration hook that the generated
// endpoint methods route through. A decorator sees the documented parameters as
// typed values, before the runtime serializes them, so it can transform a call
// without the generated code knowing anything about it.
//
// The API 二次加密和签名 layer (apisecurity.go) is the first implementation:
// rather than teaching every generated call site that an endpoint supports
// encryption, the caller installs the decorator on the client and names the
// endpoints it covers.

// RequestInfo identifies the call being decorated. It is filled by the runtime
// from the client, so a decorator sees the account the call is actually made
// for, including a client derived with WithCredentials.
type RequestInfo struct {
	// URLPath is the full request URL with the query removed: scheme, host and
	// path. It is the string WeChat's request signature and GCM additional
	// authenticated data cover.
	URLPath string
	// Path is the endpoint path alone, e.g. "/wxa/getuserriskrank".
	Path string
	// Method is the HTTP verb of the call.
	Method string
	// Credentials are the account credentials of the client issuing the call.
	Credentials Credentials
}

// ModifiedRequest is what a RequestModifier substitutes for the runtime's
// default assembly: the exact body bytes to send plus any extra HTTP headers.
type ModifiedRequest struct {
	// Body is the bytes to send as the request body.
	Body []byte
	// Headers are added to the request (a name already set by the runtime is
	// overwritten).
	Headers map[string]string
}

// RequestModifier decorates the outgoing request and the incoming response of an
// endpoint call. Both methods are consulted for every call; returning a nil
// *ModifiedRequest (or handled=false) means "leave this to the runtime".
//
// Implementations must be safe for concurrent use, because one client serves
// concurrent calls.
type RequestModifier interface {
	// ModifyRequest returns the decorated request, or nil to let the runtime
	// assemble the call normally.
	ModifyRequest(info RequestInfo, req any) (*ModifiedRequest, error)
	// ModifyResponse may replace a response body before the runtime decodes it.
	// handled=false keeps the original body. Returning an error fails the call,
	// which is how a decorator rejects a response it cannot accept (for example
	// one whose signature does not verify).
	ModifyResponse(info RequestInfo, header http.Header, body []byte) (decoded []byte, handled bool, err error)
}

// ErrModifierRejected reports a response a decorator refused to accept.
var ErrModifierRejected = errors.New("wechat: request modifier rejected the response")

// Use installs request modifiers on the client, in the order given: each is
// asked to modify the request until one returns a decorated request, and every
// one is asked to modify the response. Install them before the client is used
// concurrently.
//
// This is the decorator entry point; platform clients expose it through their
// embedded EndpointClient, so a caller can do:
//
//	client := miniprogram.New(miniprogram.Config{AppID: ..., AppSecret: ...})
//	client.Use(apiSecurity)
func (c *EndpointClient) Use(mods ...RequestModifier) {
	c.modifiers = append(c.modifiers, mods...)
}

// modifyRequest asks the installed decorators for a decorated request, returning
// the first one that is accepted.
func (c *EndpointClient) modifyRequest(info RequestInfo, req any) (*ModifiedRequest, error) {
	if len(c.modifiers) == 0 || req == nil {
		return nil, nil
	}
	for _, m := range c.modifiers {
		modified, err := m.ModifyRequest(info, req)
		if err != nil {
			return nil, err
		}
		if modified != nil {
			return modified, nil
		}
	}
	return nil, nil
}

// modifyResponse lets every decorator inspect the response body in order,
// chaining replacements.
func (c *EndpointClient) modifyResponse(info RequestInfo, header http.Header, body []byte) ([]byte, error) {
	for _, m := range c.modifiers {
		decoded, handled, err := m.ModifyResponse(info, header, body)
		if err != nil {
			return nil, err
		}
		if handled {
			body = decoded
		}
	}
	return body, nil
}
