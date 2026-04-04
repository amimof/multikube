MODULE   = $(shell env GO111MODULE=on $(GO) list -m)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOVERSION=$(shell go version | awk -F\go '{print $$3}' | awk '{print $$1}')
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
TESTPKGS = $(shell env GO111MODULE=on $(GO) list -f \
			'{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' \
			$(PKGS))
BUILDPATH ?= $(BIN)
SRC_FILES=find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.cache/*" -print0 | xargs -0 
BIN      = $(CURDIR)/bin
TBIN		 = $(CURDIR)/test/bin
INTDIR	 = $(CURDIR)/test/int-test
API_SERVICES=./api/services
GO			 = go
TIMEOUT  = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m➜\033[0m")
MAKEFLAGS += -j2
RUNTIME ?= "nerdctl"

export GO111MODULE=on
export CGO_ENABLED=0

.PHONY: all
all: server 

.PHONY: server
server: | $(BIN) ; $(info $(M) building server executable to $(BUILDPATH)/$(BINARY_NAME)) @ ## Build program binary
	$Q $(GO) build \
		-tags release \
		-ldflags '-X main.VERSION=${VERSION} -X main.DATE=${DATE} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -X main.GOVERSION=${GOVERSION}' \
		-o $(BUILDPATH)/$${BINARY_NAME:=multikube} cmd/multikube/main.go

cli: | $(BIN) ; $(info $(M) building client executable to $(BUILDPATH)/$(BINARY_NAME)) @ ## Build program binary
	$Q $(GO) build \
		-tags release \
		-ldflags '-X main.VERSION=${VERSION} -X main.DATE=${DATE} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -X main.GOVERSION=${GOVERSION}' \
		-o $(BUILDPATH)/$${BINARY_NAME:=multikubectl} ./cmd/multikubectl/

.PHONY: oci
oci: ; $(info $(M) building container image) @ ## Build container image from Dockerfile
	$(RUNTIME) build -t ghcr.io/amimof/multikube:${VERSION} .
	$(RUNTIME) tag ghcr.io/amimof/multikube:${VERSION} ghcr.io/amimof/multikube:latest

.PHONY: protos
protos: ; $(info $(M) generating protos) @ ## Generate protos
	buf generate

# Tools

$(BIN):
	@mkdir -p $(BIN)
$(TBIN):
	@mkdir -p $@
$(INTDIR):
	@mkdir -p $@
$(TBIN)/%: | $(TBIN) ; $(info $(M) building $(PACKAGE))
	$Q tmp=$$(mktemp -d); \
	   env GOBIN=$(TBIN) $(GO) install $(PACKAGE) \
		|| ret=$$?; \
	   #rm -rf $$tmp ; exit $$ret

GOCILINT = $(TBIN)/golangci-lint
$(TBIN)/golangci-lint: PACKAGE=github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0

GOLICENSES = $(TBIN)/go-licenses
$(TBIN)/go-licenses: PACKAGE=github.com/google/go-licenses/v2@v2.0.1

# Tests
.PHONY: lint
lint: | $(GOCILINT) ; $(info $(M) running golangci-lint) @ ## Runs static code analysis using golangci-lint
	$Q $(GOCILINT) run --timeout=5m

.PHONY: test
test: ; $(info $(M) running go test) @ ## Runs unit tests
	$Q $(GO) test -count=1 -v ./...

.PHONY: fmt
fmt: ; $(info $(M) running gofmt) @ ## Formats Go code
	$Q $(GO) fmt ./...

.PHONY: vet
vet: ; $(info $(M) running go vet) @ ## Examines Go source code and reports suspicious constructs, such as Printf calls whose arguments do not align with the format string
	$Q $(GO) vet ./...

.PHONY: race
race: ; $(info $(M) running go race) @ ## Runs tests with data race detection
	$Q CGO_ENABLED=1 $(GO) test -race -short ./...

.PHONY: benchmark
benchmark: ; $(info $(M) running go benchmark test) @ ## Benchmark tests to examine performance
	$Q $(GO) test -run=__absolutelynothing__ -bench=. $(PKGS)

.PHONY: coverage
coverage: ; $(info $(M) running go coverage) @ ## Runs tests and generates code coverage report at ./test/coverage.out
	$Q mkdir -p $(CURDIR)/test/
	$Q $(GO) test -coverprofile="$(CURDIR)/test/coverage.out" ./...
	$Q $(GO) tool cover -html "$(CURDIR)/test/coverage.out" -o "$(CURDIR)/test/coverage.html"

.PHONY: checkfmt
checkfmt: ; $(info $(M) running checkfmt) @ ## Checks if code is formatted with go fmt and errors out if not
	@test "$(shell $(SRC_FILES) gofmt -l)" = "" \
    || { echo "Code not formatted, please run 'make fmt'"; exit 2; }

.PHONY: license-report
license-report: | $(GOLICENSES) ## Analyzes go dependencies and prints the result as CSV
	@echo "$(M) running license report"
	$Q $(GOLICENSES) report ./...

.PHONY: license-check
license-check: | $(GOLICENSES) ## Checks whether licenses for a package are not allowed
	@echo "$(M) running license check"
	$Q $(GOLICENSES) check ./... --allowed_licenses="Unlicense,ISC,MPL-2.0,BSD-2-Clause,BSD-3-Clause,MIT,Apache-2.0"

.PHONY: license
license: ## Runs license-check, license-report and license-save
	@echo "$(M) running license targets"
	$Q $(MAKE) license-check
	$Q $(MAKE) license-report


# Misc

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m∙ %s:\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:	## Print version information
	@echo App: $(VERSION)
	@echo Go: $(GOVERSION)
	@echo Commit: $(COMMIT)
	@echo Branch: $(BRANCH)

.PHONY: clean
clean: ; $(info $(M) cleaning)	@ ## Cleanup everything
	@rm -rfv $(BIN)
	@rm -rfv $(TBIN)
	@rm -rfv $(CURDIR)/test
