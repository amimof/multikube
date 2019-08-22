# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

BINARY=multikube
GOARCH=amd64
VERSION=1.0.0-alpha.7
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOVERSION=$(shell go version | awk -F\go '{print $$3}' | awk '{print $$1}')
GITHUB_USERNAME=amimof
PKG_LIST=$$(go list ./... | grep -v /vendor/)
SRC_FILES=find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.cache/*" -print0 | xargs -0 
PROJ_FILES=find . -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.cache/*" -print0 | xargs -0 
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -X main.GOVERSION=${GOVERSION}"

# Build the project
all: build

dep:
	GO111MODULES=on go get -v -d ./cmd/multikube/... ; \
	go get -u github.com/fzipp/gocyclo; \
	go get -u golang.org/x/lint/golint; \
	go get github.com/gordonklaus/ineffassign; \
	go get -u github.com/client9/misspell/cmd/misspell; \

fmt:
	$(SRC_FILES) gofmt -s -e -d -w; \

vet:
	go vet ${PKG_LIST}; \

race:
	go test -race -short ${PKG_LIST}; \

msan:
	go test -msan -short ${PKG_LIST}; \

golint:
	${GOPATH}/bin/golint -set_exit_status ${PKG_LIST}; \

gocyclo:
	$(SRC_FILES) ${GOPATH}/bin/gocyclo -over 30; \

ineffassign:
	$(SRC_FILES) ${GOPATH}/bin/ineffassign; \

misspell:
	$(PROJ_FILES) ${GOPATH}/bin/misspell; \

checkfmt:
	if [ "`$(SRC_FILES) gofmt -l`" != "" ]; then \
		echo "Code not formatted, please run 'make fmt'"; \
		exit 1; \
	fi

ci: fmt vet race msan gocyclo golint ineffassign misspell 

test:
	go test ; \


linux: dep
	mkdir -p ./bin/; \
	GO111MODULES=on CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${BINARY}-linux-${GOARCH} ./cmd/multikube/...

rpi: dep
	mkdir -p ./bin/; \
	GO111MODULES=on CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LDFLAGS} -o ./bin/${BINARY}-linux-arm ./cmd/multikube/...

darwin: dep
	mkdir -p ./bin/; \
	GO111MODULES=on CGO_ENABLED=0 GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${BINARY}-darwin-${GOARCH} ./cmd/multikube/...

windows: dep
	mkdir -p ./bin/; \
	GO111MODULES=on CGO_ENABLED=0 GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${BINARY}-windows-${GOARCH}.exe ./cmd/multikube/...

docker_build:
	docker run --rm -v "${PWD}":/go/src/gitlab.com/amimof/multikube -w /go/src/gitlab.com/amimof/multikube golang:${GOVERSION} make fmt test
	docker build -t registry.gitlab.com/amimof/multikube:${VERSION} .
	docker tag registry.gitlab.com/amimof/multikube:${VERSION} registry.gitlab.com/amimof/multikube:latest

docker_push:
	docker push registry.gitlab.com/amimof/multikube:${VERSION}
	docker push registry.gitlab.com/amimof/multikube:latest

docker: docker_build docker_push

build: linux darwin rpi windows

clean:
	-rm -rf ./bin/

.PHONY: linux darwin windows test fmt clean