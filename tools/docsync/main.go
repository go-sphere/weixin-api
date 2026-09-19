// Command docsync keeps the generated WeChat clients in step with the Chinese
// server-API documentation trees this repository implements.
//
// It runs as three stages, so the expensive network crawl is not repeated while
// the extractor or the generator is being iterated on:
//
//	crawl    fetch every page breadth-first, normalize the body (the "content"
//	         block only — the shared sidebar and site chrome are excluded, so
//	         reordering the TOC does not mark every page as changed), record a
//	         SHA-256 hash of it in a committed manifest, and mirror the raw
//	         bytes into a gitignored local page cache.
//	swagger  rebuild the OpenAPI contracts, offline, from the cached pages.
//	gen      rebuild the Go client packages, offline, from the cached pages.
//
// Diffing the manifest after a crawl answers "which documentation pages changed
// upstream since the last run" — the input for deciding whether the clients
// need updating.
//
// Usage (from anywhere inside the repository; the tool locates the repo root by
// walking up to the go.mod declaring the library module):
//
//	docsync                  all three stages (crawl, swagger, gen)
//	docsync -stage=crawl     crawl only: refresh the page cache + manifest
//	docsync -stage=swagger   swagger only, from the local cache (no network)
//	docsync -stage=gen       Go code only, from the local cache (no network)
//	docsync -check           crawl and report drift only; exit status 1 on drift
//	docsync -format=json
//
// Set IGNORE_CACHE=1 (or pass -ignore-cache) to make an offline stage re-crawl
// from the network first.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

const (
	exitOK    = 0
	exitDrift = 1 // -check found upstream changes
	exitError = 2
)

// Pipeline stages. The crawl is the only stage that needs the network and the
// only one whose input is not part of the repository, so it is separated: it
// mirrors every page into the gitignored local cache, and the later stages
// rebuild their artifacts from that cache without crawling.
const (
	stageAll     = "all"     // crawl, then swagger, then Go code
	stageCrawl   = "crawl"   // pages -> local cache + manifest
	stageSwagger = "swagger" // cached pages -> OpenAPI contracts
	stageGen     = "gen"     // cached pages -> Go client packages
)

// repoModule is the module path of the enclosing library; repoRoot walks up
// until it finds the go.mod declaring it.
const repoModule = "github.com/go-sphere/weixin-api"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		check        = flag.Bool("check", false, "report drift against the manifest without updating anything; exit 1 when drift is found")
		stage        = flag.String("stage", stageAll, "pipeline stage: "+stageAll+", "+stageCrawl+", "+stageSwagger+" or "+stageGen)
		ignoreCache  = flag.Bool("ignore-cache", envBool("IGNORE_CACHE"), "re-crawl from the network instead of loading the local page cache (env IGNORE_CACHE)")
		format       = flag.String("format", "text", "report format: text or json")
		manifestPath = flag.String("manifest", "tools/docsync/manifest.json", "manifest path, relative to the repository root")
		cacheDir     = flag.String("cache-dir", ".docs-cache", "local page cache directory, relative to the repository root; gitignored and safe to delete")
		swaggerDir   = flag.String("swagger-dir", "tools/docsync/swagger", "OpenAPI contract output directory, relative to the repository root")
		genDir       = flag.String("gen-dir", ".", "output root for the generated packages, relative to the repository root; the default writes into the library module package directories (miniprogram/, official/)")
		llmFallback  = flag.Bool("llm", false, "use the LLM (OpenAI-compatible endpoint from LLM.ENV) to extract pages the deterministic parser could not")
		workers      = flag.Int("workers", 8, "concurrent fetch workers")
		timeout      = flag.Duration("timeout", 45*time.Second, "per-request timeout")
	)
	flag.Parse()

	if *stage != stageAll && *stage != stageCrawl && *stage != stageSwagger && *stage != stageGen {
		fmt.Fprintf(os.Stderr, "docsync: unknown -stage %q; want %s, %s, %s or %s\n", *stage, stageAll, stageCrawl, stageSwagger, stageGen)
		return exitError
	}
	current := *stage
	if *check {
		// Drift detection is defined against a live crawl, and it must not
		// touch the working tree.
		current = stageCrawl
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "docsync:", err)
		return exitError
	}
	manifestFile := resolvePath(root, *manifestPath)
	cacheRoot := resolvePath(root, *cacheDir)
	swaggerDirPath := resolvePath(root, *swaggerDir)
	genRoot := resolvePath(root, *genDir)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	old, err := loadManifest(manifestFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docsync:", err)
		return exitError
	}

	// Crawl when the stage asks for it or the cache was explicitly ignored;
	// otherwise the pages are loaded from the local cache and no network is
	// used. -check reports drift and leaves the working tree alone, so it
	// crawls without a cache to write into.
	var (
		pages map[string]*page
		drift bool
	)
	if current == stageAll || current == stageCrawl || *ignoreCache {
		var cache *pageCache
		if !*check {
			cache = newPageCache(cacheRoot)
		}
		pages, drift, err = crawlStage(ctx, &http.Client{Timeout: *timeout}, *workers, old, *format, cache)
		if err != nil {
			fmt.Fprintln(os.Stderr, "docsync:", err)
			return exitError
		}
		if !*check {
			if err := cache.Prune(); err != nil {
				fmt.Fprintln(os.Stderr, "docsync:", err)
				return exitError
			}
			// The cache and the manifest are both produced by the crawl, so
			// they are written together and stay consistent with each other.
			if err := saveManifest(manifestFile, pages); err != nil {
				fmt.Fprintln(os.Stderr, "docsync:", err)
				return exitError
			}
		}
	} else {
		pages, err = loadCachedPages(cacheRoot, old)
		if err != nil {
			fmt.Fprintln(os.Stderr, "docsync:", err)
			return exitError
		}
		fmt.Printf("docsync: loaded %d cached page(s) from %s (IGNORE_CACHE=1 to re-crawl)\n", len(pages), cacheRoot)
	}

	ops, err := collectOperations(pages)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docsync:", err)
		return exitError
	}
	if *llmFallback {
		ops, err = refineWithLLM(ctx, pages, ops)
		if err != nil {
			fmt.Fprintln(os.Stderr, "docsync: llm fallback:", err)
			return exitError
		}
	}
	assignOperationIDs(ops)

	if *check {
		fmt.Printf("%d API operations extracted\n", len(ops))
		// The LLM fallback is not reproducible, so a run that needed it can
		// never be reported as "no drift".
		if drift || *llmFallback {
			return exitDrift
		}
		return exitOK
	}
	if current == stageCrawl {
		fmt.Printf("crawl: %d page(s) -> %s\nmanifest: %s\n", len(pages), cacheRoot, manifestFile)
		return exitOK
	}

	swaggerFiles := make(map[string][]byte)
	for _, pkg := range packageOperations(ops) {
		if current == stageAll || current == stageSwagger {
			jsonDoc, yamlDoc, buildErr := buildSwagger(ctx, pkg.pkg.Title, pkg.Ops)
			if buildErr != nil {
				fmt.Fprintln(os.Stderr, "docsync:", buildErr)
				return exitError
			}
			swaggerFiles[pkg.pkg.Dir+".swagger.json"] = jsonDoc
			swaggerFiles[pkg.pkg.Dir+".swagger.yaml"] = yamlDoc
		}
		if current == stageAll || current == stageGen {
			target := filepath.Join(genRoot, pkg.pkg.Dir)
			if err := generateGoClient(target, pkg.pkg.Name, pkg.Ops); err != nil {
				fmt.Fprintln(os.Stderr, "docsync: generate:", err)
				return exitError
			}
			fmt.Printf("package %s: %d operations -> %s\n", pkg.pkg.Name, len(pkg.Ops), target)
		}
	}
	if len(swaggerFiles) > 0 {
		if err := writeSwagger(swaggerDirPath, swaggerFiles); err != nil {
			fmt.Fprintln(os.Stderr, "docsync:", err)
			return exitError
		}
		fmt.Printf("swagger: %d operations -> %s\n", len(ops), swaggerDirPath)
	}

	fmt.Printf("operations: %d\nmanifest:   %s\ncache:      %s\n", len(ops), manifestFile, cacheRoot)
	return exitOK
}

// crawlStage performs the network crawl, reports drift against the previous
// manifest and, when cache is non-nil, mirrors every page into the local page
// cache. It returns the fetched pages and whether they differ from old.
func crawlStage(ctx context.Context, client *http.Client, workers int, old *Manifest, format string, cache *pageCache) (map[string]*page, bool, error) {
	tracked := make(map[string]bool, len(old.Pages))
	for key := range old.Pages {
		tracked[key] = true
	}
	// A nil *pageCache assigned to the interface would still compare non-nil
	// and be called, so the conversion has to be explicit.
	var sink pageSink
	if cache != nil {
		sink = cache
	}
	pages, fatal, skipped, err := crawl(ctx, client, workers, tracked, sink)
	if err != nil {
		return nil, false, err
	}
	for _, s := range skipped {
		fmt.Fprintf(os.Stderr, "docsync: skipping untracked page (broken link on a doc page): %s: %v\n", s.url, s.err)
	}
	if len(fatal) > 0 {
		fmt.Fprintf(os.Stderr, "docsync: %d tracked pages failed to fetch; refusing to update the manifest (the page cache may be partial):\n", len(fatal))
		for _, f := range fatal {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", f.url, f.err)
		}
		return nil, false, fmt.Errorf("%d tracked page(s) failed to fetch", len(fatal))
	}
	rep := buildReport(old, pages)
	printReport(os.Stdout, format, rep)
	return pages, rep.drift(), nil
}

// envBool reports whether the named environment variable is set to a true-like
// value. An unset or empty variable is false, as are "0", "false" and "no".
func envBool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "", "0", "false", "no":
		return false
	default:
		return true
	}
}

// resolvePath resolves p against the repository root, leaving absolute paths
// untouched.
func resolvePath(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

// repoRoot walks up from the working directory to the go.mod that declares
// the library module (the nested tool module also has a go.mod, so "nearest
// go.mod" is not enough) and returns that directory.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if module := modulePath(filepath.Join(dir, "go.mod")); module == repoModule {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod declaring module %q found above the working directory; run inside the repository", repoModule)
		}
		dir = parent
	}
}

// modulePath returns the module directive of the go.mod at path, or "" when
// the file is missing or unparsable.
func modulePath(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if module, ok := strings.CutPrefix(line, "module "); ok {
			return strings.Trim(strings.TrimSpace(module), `"`)
		}
	}
	return ""
}
