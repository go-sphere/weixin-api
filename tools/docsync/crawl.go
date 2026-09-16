package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// seedURLs are the roots of the Chinese documentation trees mirrored by this
// tool. Every WeChat doc page renders its document body inside a
// div#docContent > div.content.custom block and the full page tree in a
// sidebar, so the seeds are enough to discover every page in scope.
//
// Note: the Mini Program server API tree lives under /miniprogram/dev/server/
// API/ in the current doc system; the old /miniprogram/dev/api-backend/ URL is
// only a legacy alias whose in-page links are broken.
var seedURLs = []string{
	"https://developers.weixin.qq.com/miniprogram/dev/server/API/",
	"https://developers.weixin.qq.com/doc/service/api/",
	"https://developers.weixin.qq.com/doc/subscription/api/",
}

const (
	userAgent    = "weixin-api-docsync (+https://github.com/go-sphere/weixin-api)"
	maxPageBytes = 4 << 20 // plenty for the largest doc page; guards against runaways
	fetchTries   = 3
)

var errNotFound = errors.New("not found")

// page is one crawled documentation page.
type page struct {
	Canonical string       // canonical URL, the manifest key
	Title     string       // first h1 of the body, else the document title
	Body      []byte       // normalized document body (TOC and site chrome excluded)
	Hash      string       // sha256 hex of Body
	Ops       []*Operation // semantic API definitions extracted from the page
}

// pageSink receives every page the crawl fetches successfully, keyed by its
// canonical URL. The crawl stage uses it to mirror the raw bytes into the local
// cache; -check passes nil so a drift report never touches the working tree.
type pageSink interface {
	Store(canonical string, raw []byte) error
}

type fetchFailure struct {
	url string
	err error
}

// crawl fetches every page reachable from seedURLs whose canonical URL stays
// inside one of the seed path prefixes, round by round, until no new pages
// appear. It returns the fetched pages (keyed by canonical URL). Failures on
// tracked URLs come back as fatal errors (a tracked page disappearing needs
// attention); failures on newly discovered URLs — typically broken editorial
// links on doc pages — come back as skips so one bad link does not abort the
// whole run. A non-nil sink receives each successfully fetched page.
func crawl(ctx context.Context, client *http.Client, concurrency int, tracked map[string]bool, sink pageSink) (map[string]*page, []fetchFailure, []fetchFailure, error) {
	prefixes, err := seedPrefixes()
	if err != nil {
		return nil, nil, nil, err
	}

	frontier := slices.Clone(seedURLs)
	seen := make(map[string]bool, 1024)
	pages := make(map[string]*page, 1024)
	var fatal, skipped []fetchFailure

	for round := 0; len(frontier) > 0; round++ {
		outcomes := make([]*fetchOutcome, len(frontier))
		var wg sync.WaitGroup
		sem := make(chan struct{}, max(1, concurrency))
		for i, raw := range frontier {
			wg.Go(func() {
				sem <- struct{}{}
				defer func() { <-sem }()
				outcomes[i] = fetchPage(ctx, client, raw, prefixes, sink)
			})
		}
		wg.Wait()

		var next []string
		for i, out := range outcomes {
			if out == nil {
				return nil, nil, nil, fmt.Errorf("round %d: fetch of %s never ran", round, frontier[i])
			}
			if out.err != nil {
				f := fetchFailure{url: frontier[i], err: out.err}
				if tracked[frontier[i]] {
					fatal = append(fatal, f)
				} else {
					skipped = append(skipped, f)
				}
				continue
			}
			pages[out.page.Canonical] = out.page
			for _, link := range out.links {
				if !seen[link] {
					seen[link] = true
					next = append(next, link)
				}
			}
		}
		frontier = next
	}
	return pages, fatal, skipped, ctx.Err()
}

func seedPrefixes() ([]string, error) {
	prefixes := make([]string, 0, len(seedURLs))
	for _, raw := range seedURLs {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse seed %s: %w", raw, err)
		}
		prefixes = append(prefixes, u.Path)
	}
	return prefixes, nil
}

func inScope(prefixes []string, path string) bool {
	return slices.ContainsFunc(prefixes, func(p string) bool {
		return strings.HasPrefix(path, p)
	})
}

// isDocPath reports whether path looks like a documentation page rather than
// a linked resource (template.zip downloads, images, …) that some doc pages
// reference.
func isDocPath(path string) bool {
	seg := path[strings.LastIndexByte(path, '/')+1:]
	dot := strings.LastIndexByte(seg, '.')
	if dot < 0 {
		return true
	}
	switch strings.ToLower(seg[dot:]) {
	case ".html", ".htm":
		return true
	default:
		return false
	}
}

type fetchOutcome struct {
	page  *page
	links []string
	err   error
}

// fetchPage downloads raw (retrying transient failures, falling back to the
// pre-canonicalization URL on 404), then parses the body and outgoing links.
// A non-nil sink receives the validated raw bytes before parsing, so the cache
// holds exactly what parsePage consumed.
func fetchPage(ctx context.Context, client *http.Client, raw string, prefixes []string, sink pageSink) *fetchOutcome {
	target := raw
	body, finalURL, err := fetch(ctx, client, target)
	if errors.Is(err, errNotFound) && strings.HasSuffix(target, ".html") {
		// The canonical key may have gained a ".html" suffix that the raw
		// page does not actually live at; retry the suffix-less form.
		body, finalURL, err = fetch(ctx, client, strings.TrimSuffix(target, ".html"))
	}
	if err != nil {
		return &fetchOutcome{err: err}
	}
	// some upstream pages contain invalid UTF-8; normalize at the source so
	// every derived artifact (cache, hash) is well-formed
	body = bytes.ToValidUTF8(body, []byte("\uFFFD"))

	title, normalized, links, contentNode, err := parsePage(body, finalURL)
	if err != nil {
		return &fetchOutcome{err: fmt.Errorf("parse %s: %w", finalURL, err)}
	}

	canonical, ok := canonicalURL(finalURL)
	if !ok || !inScope(prefixes, mustPath(canonical)) {
		return &fetchOutcome{err: fmt.Errorf("%s redirected out of scope to %s", raw, finalURL)}
	}

	if sink != nil {
		if err := sink.Store(canonical, body); err != nil {
			return &fetchOutcome{err: fmt.Errorf("cache %s: %w", canonical, err)}
		}
	}

	// Semantic extraction: the unified doc template (调用方式/请求参数/返回参数/
	// 错误码) is parsed deterministically into API operations. Catalogue
	// pages and non-standard pages yield none.
	p := &page{
		Canonical: canonical,
		Title:     title,
		Body:      normalized,
		Hash:      hashBody(normalized),
	}
	p.Ops = parseOperations(canonical, title, contentNode)

	var inScopeLinks []string
	for _, link := range links {
		if inScope(prefixes, mustPath(link)) && isDocPath(mustPath(link)) {
			inScopeLinks = append(inScopeLinks, link)
		}
	}
	return &fetchOutcome{page: p, links: inScopeLinks}
}

func fetch(ctx context.Context, client *http.Client, raw string) ([]byte, string, error) {
	var lastErr error
	for attempt := range fetchTries {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(time.Duration(attempt) * 750 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("User-Agent", userAgent)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		switch {
		case resp.StatusCode == http.StatusOK:
			data, err := io.ReadAll(io.LimitReader(resp.Body, maxPageBytes))
			resp.Body.Close()
			if err != nil {
				lastErr = err
				continue
			}
			return data, resp.Request.URL.String(), nil
		case resp.StatusCode == http.StatusNotFound:
			resp.Body.Close()
			return nil, resp.Request.URL.String(), fmt.Errorf("%w (%s)", errNotFound, raw)
		default:
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
	}
	return nil, "", lastErr
}

// canonicalURL normalizes u into the manifest key: same-host doc URLs only,
// no query or fragment, and ".html" appended when the last path segment has
// no extension (the site serves /a/b and /a/b.html as identical pages, and
// both forms appear in links; without this, the same page would show up as
// two manifest entries and flip between them).
func canonicalURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if u.Host != "developers.weixin.qq.com" {
		return "", false
	}
	u.Scheme = "https"
	u.User = nil
	u.Fragment = ""
	u.RawQuery = ""
	u.ForceQuery = false
	if u.Path == "" {
		u.Path = "/"
	}
	if !strings.HasSuffix(u.Path, "/") {
		seg := u.Path[strings.LastIndexByte(u.Path, '/')+1:]
		if !strings.Contains(seg, ".") {
			u.Path += ".html"
		}
	}
	return u.String(), true
}

func mustPath(canonical string) string {
	u, err := url.Parse(canonical)
	if err != nil {
		return ""
	}
	return u.Path
}

// parsePage extracts the title, normalized body markup, all outgoing anchor
// URLs (canonicalized) and the raw content node of a rendered doc page.
func parsePage(data []byte, baseURL string) (title string, body []byte, links []string, content *html.Node, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return "", nil, nil, nil, err
	}
	contentSel := doc.Find("#docContent .content.custom").First()
	if contentSel.Length() == 0 {
		return "", nil, nil, nil, errors.New(`document content block div.content.custom not found`)
	}
	content = contentSel.Nodes[0]

	body = normalize(content)
	title = firstHeading(content, "h1")
	// h1 headings render as "<h1><a class=header-anchor>#</a> Title</h1>";
	// drop the anchor's own "#" from the extracted title.
	title = strings.TrimLeft(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(title), "#")), " ")
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", nil, nil, nil, err
	}
	seen := make(map[string]bool)
	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		ref, err := url.Parse(href)
		if err != nil {
			return
		}
		resolved := base.ResolveReference(ref)
		canonical, ok := canonicalURL(resolved.String())
		if ok && !seen[canonical] {
			seen[canonical] = true
			links = append(links, canonical)
		}
	})
	return strings.TrimSpace(title), body, links, content, nil
}

func hashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// normalize serializes the content block into a stable byte stream: tags with
// content-bearing attributes only (href/src/colspan/rowspan), comments and
// script/style dropped, whitespace collapsed. Volatile attributes (Vue scoped
// data-v hashes, classes, inline styles) are stripped so redeployment churn
// does not show up as content drift, while text, structure and links are kept.
func normalize(n *html.Node) []byte {
	var buf bytes.Buffer
	writeNormalized(&buf, n)
	return bytes.TrimSpace(buf.Bytes())
}

func writeNormalized(buf *bytes.Buffer, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		if text := strings.Join(strings.Fields(n.Data), " "); text != "" {
			buf.WriteByte(' ')
			buf.WriteString(text)
		}
	case html.ElementNode:
		switch strings.ToLower(n.Data) {
		case "script", "style", "template", "noscript":
			return
		}
		buf.WriteByte('<')
		buf.WriteString(strings.ToLower(n.Data))
		for _, attr := range n.Attr {
			switch strings.ToLower(attr.Key) {
			case "href", "src", "colspan", "rowspan":
				fmt.Fprintf(buf, ` %s=%q`, strings.ToLower(attr.Key), attr.Val)
			}
		}
		buf.WriteByte('>')
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNormalized(buf, c)
		}
		buf.WriteString("</")
		buf.WriteString(strings.ToLower(n.Data))
		buf.WriteByte('>')
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeNormalized(buf, c)
		}
	}
}

// firstHeading returns the concatenated text of the first heading of the
// given level inside n, or "" when the body has none (index pages).
func firstHeading(n *html.Node, level string) string {
	if n.Type == html.ElementNode && strings.EqualFold(n.Data, level) {
		return headingText(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if text := firstHeading(c, level); text != "" {
			return text
		}
	}
	return ""
}

func headingText(n *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			if text := strings.Join(strings.Fields(n.Data), " "); text != "" {
				parts = append(parts, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(parts, " ")
}
