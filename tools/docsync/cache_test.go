package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cacheRoundTripDoc is a minimal but complete doc page: it has the content
// block parsePage requires plus the 调用方式/请求参数 template the extractor
// matches, so a cached copy yields the same operation as a live parse.
const cacheRoundTripDoc = `<!DOCTYPE html><html><body><div id="docContent"><div class="content custom">
<h1>获取接口调用凭据</h1>
<h2>1. 调用方式</h2><h3>HTTPS 调用</h3>
<p>GET https://api.weixin.qq.com/cgi-bin/token?appid=APPID&amp;secret=SECRET</p>
<h2>2. 请求参数</h2><h3>请求体 Request Payload</h3>
<table><thead><tr><th>参数名</th><th>类型</th><th>必填</th><th>说明</th></tr></thead>
<tbody><tr><td>grant_type</td><td>string</td><td>是</td><td>凭证类型</td></tr></tbody></table>
</div></div></body></html>`

// cacheManifest is a one-page manifest whose hash matches cacheRoundTripDoc.
func cacheManifest(t *testing.T, canonical string, raw []byte) *Manifest {
	t.Helper()
	title, normalized, _, content, err := parsePage(raw, canonical)
	if err != nil {
		t.Fatalf("parsePage: %v", err)
	}
	if ops := parseOperations(canonical, title, content); len(ops) != 1 {
		t.Fatalf("fixture must yield exactly one operation, got %d", len(ops))
	}
	return &Manifest{
		Algorithm: manifestAlgorithm,
		Seeds:     seedURLs,
		Pages:     map[string]PageRecord{canonical: {Hash: hashBody(normalized), Title: title}},
	}
}

const cacheCanonical = "https://developers.weixin.qq.com/doc/service/api/base/api_getaccesstoken.html"

// TestPageCacheRoundTrip proves the offline path is equivalent to a live
// parse: bytes stored by the crawl sink load back into a page whose normalized
// body, hash, title and extracted operation match the original.
func TestPageCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(cacheRoundTripDoc)
	m := cacheManifest(t, cacheCanonical, raw)

	cache := newPageCache(dir)
	if err := cache.Store(cacheCanonical, raw); err != nil {
		t.Fatalf("Store: %v", err)
	}
	// Store must write the raw bytes verbatim, not a re-serialization.
	full, err := cache.path(cacheCanonical)
	if err != nil {
		t.Fatal(err)
	}
	onDisk, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("cached file missing: %v", err)
	}
	if string(onDisk) != string(raw) {
		t.Error("cache did not store the raw bytes verbatim")
	}

	pages, err := loadCachedPages(dir, m)
	if err != nil {
		t.Fatalf("loadCachedPages: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(pages))
	}
	got := pages[cacheCanonical]
	if got.Hash != m.Pages[cacheCanonical].Hash {
		t.Errorf("hash = %s, want %s", got.Hash, m.Pages[cacheCanonical].Hash)
	}
	if got.Title != m.Pages[cacheCanonical].Title {
		t.Errorf("title = %q, want %q", got.Title, m.Pages[cacheCanonical].Title)
	}
	if len(got.Ops) != 1 || got.Ops[0].Method != "GET" || got.Ops[0].Path != "/cgi-bin/token" {
		t.Fatalf("cached extraction differs: %+v", got.Ops)
	}
}

// TestPageCacheMissingPageFails pins the guard that makes an offline stage
// safe: an incomplete cache must be an error naming the missing page and the
// recovery command, never a silently truncated client.
func TestPageCacheMissingPageFails(t *testing.T) {
	dir := t.TempDir()
	m := cacheManifest(t, cacheCanonical, []byte(cacheRoundTripDoc))
	// nothing stored
	_, err := loadCachedPages(dir, m)
	if err == nil {
		t.Fatal("expected an error for a page missing from the cache")
	}
	for _, want := range []string{"missing 1 of 1", "make docs-crawl", "IGNORE_CACHE=1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q lacks %q", err, want)
		}
	}
}

// TestPageCacheStalePageWarnsUnchanged proves a cache entry whose content no
// longer matches the manifest is reported but does not block generation, and
// that generation from it is not silently wrong: the manifest hash is what the
// caller compares against.
func TestPageCacheStalePageWarnsUnchanged(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(cacheRoundTripDoc)
	m := cacheManifest(t, cacheCanonical, raw)

	cache := newPageCache(dir)
	// Store a different page under the same URL: it still parses, so the run
	// continues, but its hash cannot match the manifest.
	stale := strings.Replace(cacheRoundTripDoc, "凭证类型", "凭证类型X", 1)
	if err := cache.Store(cacheCanonical, []byte(stale)); err != nil {
		t.Fatal(err)
	}
	pages, err := loadCachedPages(dir, m)
	if err != nil {
		t.Fatalf("a stale page must not fail the run: %v", err)
	}
	got := pages[cacheCanonical]
	if got.Hash == m.Pages[cacheCanonical].Hash {
		t.Fatal("fixture is not actually stale: hash still matches the manifest")
	}
	if len(got.Ops) != 1 {
		t.Errorf("stale page should still yield its operation, got %d", len(got.Ops))
	}
}

// TestPageCacheUnsafePathRejected keeps a redirected page from escaping the
// cache directory.
func TestPageCacheUnsafePathRejected(t *testing.T) {
	cache := newPageCache(t.TempDir())
	if _, err := cache.path("https://developers.weixin.qq.com/../../etc/passwd"); err == nil {
		t.Fatal("expected an unsafe cache path to be rejected")
	}
}

// TestPageCachePruneDropsVanishedPages checks that a page which disappeared
// upstream does not linger in the cache.
func TestPageCachePruneDropsVanishedPages(t *testing.T) {
	dir := t.TempDir()
	cache := newPageCache(dir)
	if err := cache.Store(cacheCanonical, []byte(cacheRoundTripDoc)); err != nil {
		t.Fatal(err)
	}
	full, _ := cache.path(cacheCanonical)
	if _, err := os.Stat(full); err != nil {
		t.Fatalf("precondition: %v", err)
	}
	// A second cache instance sees no kept pages, standing in for a crawl that
	// no longer returns this URL.
	if err := newPageCache(dir).Prune(); err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if _, err := os.Stat(full); !os.IsNotExist(err) {
		t.Errorf("stale cache file survived Prune: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(full)); err != nil {
		t.Errorf("Prune removed the directory tree: %v", err)
	}
}
