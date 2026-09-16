package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const manifestAlgorithm = "sha256"

// PageRecord is the per-page manifest entry: the hash of the normalized
// document body plus the page title for human-readable diff review.
type PageRecord struct {
	Hash  string `json:"hash"`
	Title string `json:"title"`
}

// Manifest maps canonical page URLs to their last recorded body hash. It is
// committed so that a plain git diff of the manifest shows which doc pages
// changed upstream.
type Manifest struct {
	Algorithm string                `json:"algorithm"`
	Seeds     []string              `json:"seeds"`
	Pages     map[string]PageRecord `json:"pages"`
}

func loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Manifest{Algorithm: manifestAlgorithm, Seeds: seedURLs, Pages: map[string]PageRecord{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	if m.Algorithm != "" && m.Algorithm != manifestAlgorithm {
		return nil, fmt.Errorf("manifest %s uses hash algorithm %q, want %q", path, m.Algorithm, manifestAlgorithm)
	}
	if m.Pages == nil {
		m.Pages = map[string]PageRecord{}
	}
	return &m, nil
}

func saveManifest(path string, pages map[string]*page) error {
	m := Manifest{
		Algorithm: manifestAlgorithm,
		Seeds:     seedURLs,
		Pages:     make(map[string]PageRecord, len(pages)),
	}
	for key, p := range pages {
		m.Pages[key] = PageRecord{Hash: p.Hash, Title: p.Title}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// pageCache mirrors the crawled pages into a local, gitignored directory so the
// extractor and the generator can be re-run without repeating the network
// crawl. The raw response bytes are stored verbatim, one file per page: that is
// exactly the input parsePage consumed during the crawl, so an offline run is
// byte-for-byte equivalent to a live one.
//
// Files are written as pages arrive, which keeps memory flat and lets the
// concurrent crawl store without locking. Prune then drops entries for pages
// that disappeared upstream.
type pageCache struct {
	dir  string
	mu   sync.Mutex
	keep map[string]bool
}

func newPageCache(dir string) *pageCache {
	return &pageCache{dir: dir, keep: make(map[string]bool)}
}

// Store writes one page's raw bytes under the file path its URL maps to.
func (c *pageCache) Store(canonical string, raw []byte) error {
	full, err := c.path(canonical)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(full, raw, 0o644); err != nil {
		return err
	}
	c.mu.Lock()
	c.keep[full] = true
	c.mu.Unlock()
	return nil
}

// Prune removes cached pages that the last crawl did not see.
func (c *pageCache) Prune() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return pruneStale(c.dir, c.keep)
}

// path maps a canonical page URL onto its file inside the cache.
func (c *pageCache) path(canonical string) (string, error) {
	u, err := url.Parse(canonical)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", canonical, err)
	}
	rel := cacheRelPath(u)
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("unsafe cache path %q for %s", rel, canonical)
	}
	return filepath.Join(c.dir, filepath.FromSlash(rel)), nil
}

// loadCachedPages rebuilds the crawl result from the local page cache, so the
// swagger and generation stages run without touching the network. The manifest
// names the pages that must be present and carries their content hash, so an
// incomplete cache is an error and a stale one is reported rather than silently
// producing artifacts from superseded markup.
func loadCachedPages(dir string, m *Manifest) (map[string]*page, error) {
	cache := newPageCache(dir)
	pages := make(map[string]*page, len(m.Pages))
	var missing, stale []string
	for _, canonical := range slices.Sorted(maps.Keys(m.Pages)) {
		full, err := cache.path(canonical)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			missing = append(missing, canonical)
			continue
		}
		title, normalized, _, content, err := parsePage(raw, canonical)
		if err != nil {
			return nil, fmt.Errorf("parse cached page %s: %w", canonical, err)
		}
		hash := hashBody(normalized)
		if hash != m.Pages[canonical].Hash {
			stale = append(stale, canonical)
		}
		p := &page{Canonical: canonical, Title: title, Body: normalized, Hash: hash}
		p.Ops = parseOperations(canonical, title, content)
		pages[canonical] = p
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("page cache %s is missing %d of %d page(s) (first: %s); run `make docs-crawl`, or set IGNORE_CACHE=1 to crawl from the network now",
			dir, len(missing), len(m.Pages), missing[0])
	}
	if len(stale) > 0 {
		fmt.Fprintf(os.Stderr, "docsync: warning: %d page(s) in the cache no longer match manifest.json (first: %s); run `make docs-crawl` to refresh\n",
			len(stale), stale[0])
	}
	return pages, nil
}

// writeSwagger writes the per-package OpenAPI documents (JSON and YAML,
// keyed by file base name) and prunes stale files in the directory.
func writeSwagger(dir string, files map[string][]byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keep := make(map[string]bool, len(files))
	for base, data := range files {
		full := filepath.Join(dir, base)
		keep[full] = true
		if err := os.WriteFile(full, data, 0o644); err != nil {
			return err
		}
	}
	return pruneStale(dir, keep)
}

// pruneStale removes files under dir that are not in the keep set.
func pruneStale(dir string, keep map[string]bool) error {
	var pruneErrs []error
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || keep[path] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			pruneErrs = append(pruneErrs, rmErr)
		}
		return nil
	})
	return errors.Join(append(pruneErrs, err)...)
}

// cacheRelPath turns a page URL into its path under the cache root.
func cacheRelPath(u *url.URL) string {
	p := u.Path
	if strings.HasSuffix(p, "/") {
		p += "index.html"
	}
	return strings.TrimPrefix(p, "/")
}

// report summarizes the drift between the last manifest and a fresh crawl.
type report struct {
	Total   int            `json:"total"`
	Seeds   []seedSummary  `json:"seeds"`
	New     []string       `json:"new,omitempty"`
	Changed []changedEntry `json:"changed,omitempty"`
	Removed []string       `json:"removed,omitempty"`
}

type seedSummary struct {
	Seed  string `json:"seed"`
	Pages int    `json:"pages"`
}

type changedEntry struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Old   string `json:"old_hash"`
	New   string `json:"new_hash"`
}

func (r *report) drift() bool {
	return len(r.New) > 0 || len(r.Changed) > 0 || len(r.Removed) > 0
}

func buildReport(old *Manifest, pages map[string]*page) *report {
	rep := &report{Total: len(pages)}
	for _, seed := range seedURLs {
		prefix := mustPath(seed)
		count := 0
		for key := range pages {
			if strings.HasPrefix(key, seed) || strings.HasPrefix(mustPath(key), prefix) {
				count++
			}
		}
		rep.Seeds = append(rep.Seeds, seedSummary{Seed: seed, Pages: count})
	}

	for key, p := range pages {
		prev, ok := old.Pages[key]
		switch {
		case !ok:
			rep.New = append(rep.New, key)
		case prev.Hash != p.Hash:
			rep.Changed = append(rep.Changed, changedEntry{URL: key, Title: p.Title, Old: prev.Hash, New: p.Hash})
		}
	}
	for key := range old.Pages {
		if _, ok := pages[key]; !ok {
			rep.Removed = append(rep.Removed, key)
		}
	}
	rep.New = slices.Sorted(slices.Values(rep.New))
	rep.Removed = slices.Sorted(slices.Values(rep.Removed))
	slices.SortFunc(rep.Changed, func(a, b changedEntry) int { return strings.Compare(a.URL, b.URL) })
	return rep
}

func printReport(w io.Writer, format string, rep *report) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			fmt.Fprintf(w, "docsync: marshal report: %v\n", err)
			return
		}
		data = append(data, '\n')
		w.Write(data)
	default:
		printTextReport(w, rep)
	}
}

func printTextReport(w io.Writer, rep *report) {
	for _, s := range rep.Seeds {
		fmt.Fprintf(w, "%6d pages  %s\n", s.Pages, s.Seed)
	}
	fmt.Fprintf(w, "%6d total\n", rep.Total)

	for _, key := range rep.New {
		fmt.Fprintf(w, "NEW      %s\n", key)
	}
	for _, c := range rep.Changed {
		fmt.Fprintf(w, "CHANGED  %s (%s)\n         %s… -> %s…\n", c.URL, c.Title, c.Old[:12], c.New[:12])
	}
	for _, key := range rep.Removed {
		fmt.Fprintf(w, "REMOVED  %s\n", key)
	}

	if !rep.drift() {
		fmt.Fprintln(w, "no documentation changes since the last run")
		return
	}
	fmt.Fprintf(w, "%d new, %d changed, %d removed page(s)\n", len(rep.New), len(rep.Changed), len(rep.Removed))
}
