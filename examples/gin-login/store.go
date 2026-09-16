package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Store maps WeChat identities (channel + openid) to a local user account and
// keeps a lookup table for the API tokens handed out at login. A process-local
// map is only good enough for a demo: use a database (and sign tokens with a
// real secret) in production.
type Store struct {
	mu      sync.Mutex
	users   map[string]*Account // key: channel + ":" + openid
	tokens  map[string]*Account // key: opaque bearer token
	secret  []byte
	nowFunc func() time.Time // seams for tests
}

// Channel identifies the WeChat platform a user logged in through.
type Channel string

const (
	ChannelMiniProgram     Channel = "miniprogram"      // 小程序
	ChannelOfficialAccount Channel = "official_account" // 公众号/服务号
)

// Account is a user identity linked to a WeChat openid.
type Account struct {
	Channel   Channel
	OpenID    string
	UnionID   string
	Nickname  string
	Avatar    string
	CreatedAt time.Time
}

func NewStore() *Store {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		// crypto/rand never fails on supported platforms.
		panic(err)
	}
	return &Store{
		users:   make(map[string]*Account),
		tokens:  make(map[string]*Account),
		secret:  secret,
		nowFunc: time.Now,
	}
}

// Upsert returns the stored account for (channel, openid), creating it from the
// provided profile fields on first sight. A user that first logged in through
// the Mini Program and then through the official account stays one account as
// long as both openids resolve to the same unionid.
func (s *Store) Upsert(ch Channel, openid, unionID, nickname, avatar string) *Account {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := string(ch) + ":" + openid
	if existing, ok := s.users[key]; ok {
		return existing
	}
	if unionID != "" {
		for _, other := range s.users {
			if other.UnionID == unionID {
				// Link the new openid to the existing account.
				s.users[key] = other
				return other
			}
		}
	}
	acc := &Account{
		Channel:   ch,
		OpenID:    openid,
		UnionID:   unionID,
		Nickname:  nickname,
		Avatar:    avatar,
		CreatedAt: s.nowFunc(),
	}
	s.users[key] = acc
	return acc
}

// IssueToken mints a signed bearer token for an account. The token is a random
// value plus a MAC so the demo can verify its origin without a database round
// trip. Keep the lookup map when you move to a DB-backed session store.
func (s *Store) IssueToken(acc *Account) string {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	payload := hex.EncodeToString(raw)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	token := payload + "." + hex.EncodeToString(mac.Sum(nil))
	s.mu.Lock()
	s.tokens[token] = acc
	s.mu.Unlock()
	return token
}

// LookupToken verifies the MAC and resolves a bearer token to its account.
func (s *Store) LookupToken(token string) (*Account, bool) {
	payload, mac, ok := strings.Cut(token, ".")
	if !ok {
		return nil, false
	}
	want := hmac.New(sha256.New, s.secret)
	want.Write([]byte(payload))
	if !hmac.Equal([]byte(mac), []byte(hex.EncodeToString(want.Sum(nil)))) {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	acc, ok := s.tokens[token]
	return acc, ok
}

// bearerOrCookieAuth is a Gin middleware that authenticates a request either
// through the "Authorization: Bearer <token>" header (used by the Mini Program
// flow, where no browser cookie exists) or through the wx_session cookie (used
// by the official account browser flow). It stores the *Account in the context.
func bearerOrCookieAuth(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		if after, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok {
			token = after
		} else if cookie, err := c.Cookie("wx_session"); err == nil {
			token = cookie
		}
		acc, ok := store.LookupToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("account", acc)
		c.Next()
	}
}

// currentAccount reads the *Account stored by bearerOrCookieAuth.
func currentAccount(c *gin.Context) (*Account, bool) {
	acc, ok := c.Get("account")
	if !ok {
		return nil, false
	}
	typed, ok := acc.(*Account)
	return typed, ok
}
