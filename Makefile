GO ?= go
GOLANGCI_LINT ?= golangci-lint
NILAWAY ?= nilaway

DIRECT_DEPS_TEMPLATE := {{if and (not .Main) (not .Indirect) (not .Replace)}}{{.Path}}{{end}}

# Resolve go-sphere modules straight from GitHub, bypassing the module proxy
# and its cached "@latest", which lags behind freshly pushed tags.
DIRECT_ORIGIN := GOPRIVATE=github.com/go-sphere/*

.DEFAULT_GOAL := check

.PHONY: deps-update tidy fmt test lint check docs-crawl docs-swagger docs-gen docs-sync docs-check docs-test docs-parity examples-test

deps-update:
	@GOWORK=off $(DIRECT_ORIGIN) $(GO) mod tidy; \
	deps="$$(GOWORK=off $(DIRECT_ORIGIN) $(GO) list -m -f '$(DIRECT_DEPS_TEMPLATE)' all)"; \
	if [ -n "$$deps" ]; then GOWORK=off $(DIRECT_ORIGIN) $(GO) get -u $$deps; fi
	GOWORK=off $(DIRECT_ORIGIN) $(GO) mod tidy

tidy:
	GOWORK=off $(GO) mod tidy

fmt:
	$(GO) fmt ./...
	$(GOLANGCI_LINT) fmt --no-config --enable gofmt --enable goimports

test:
	$(GO) test ./...

lint:
	$(GOLANGCI_LINT) fmt --no-config --enable gofmt --enable goimports --diff
	$(GO) vet ./...
	$(GOLANGCI_LINT) run --no-config
	$(NILAWAY) -include-pkgs="$$($(GO) list -m)" ./...

check:
	GOWORK=off $(GO) mod tidy -diff
	$(MAKE) lint
	$(MAKE) test

# docsync mirrors the WeChat doc trees in a separate Go module so its
# dependency (golang.org/x/net) stays out of the library's go.mod.
#
# The pipeline is split into stages so the network crawl — the only slow and
# non-reproducible step — runs once into a gitignored local page cache, and the
# extractor and the generator can then be iterated offline against that cache:
#
#   docs-crawl     every page -> .docs-cache/ + manifest.json      (network)
#   docs-swagger   cached pages -> swagger/*.{json,yaml}           (offline)
#   docs-gen       cached pages -> miniprogram/, official/         (offline)
#   docs-sync      all three, in that order
#
# Set IGNORE_CACHE=1 to make an offline stage re-crawl from the network first:
#
#   IGNORE_CACHE=1 make docs-gen
DOCSYNC := cd tools/docsync && $(GO) run .

docs-crawl:
	$(DOCSYNC) -stage=crawl

docs-swagger:
	$(DOCSYNC) -stage=swagger

docs-gen:
	$(DOCSYNC) -stage=gen

docs-sync:
	$(DOCSYNC) -stage=all

docs-check:
	$(DOCSYNC) -check

docs-test:
	cd tools/docsync && $(GO) test ./...

# docs-parity replays the wire behaviour captured from the hand-written
# clients against the generated ones. The expected requests live in
# tools/docsync/parity/testdata/golden.json and are authored data: no test
# rewrites them, so a regression cannot redefine its own baseline.
docs-parity:
	cd tools/docsync && $(GO) test ./parity/...

# The example is its own module; the library module's test run does not reach it.
examples-test:
	cd examples/gin-login && $(GO) test ./...
