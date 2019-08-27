# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

PROJECT=multikube
GOARCH=amd64
VERSION=$(shell git describe --tags --abbrev=0)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOVERSION=$(shell go version | awk -F\go '{print $$3}' | awk '{print $$1}')
GITHUB_USERNAME=amimof
REPO=gitlab.com/${GITHUB_USERNAME}/${PROJECT}
PKG_LIST=$$(go list ./... | grep -v /vendor/)
SRC_FILES=find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.cache/*" -print0 | xargs -0 
PROJ_FILES=find . -type f -not -path "./vendor/*" -not -path "./.git/*" -not -path "./.cache/*" -print0 | xargs -0 
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -X main.GOVERSION=${GOVERSION}"
DOCKER_REGISTRY=registry.gitlab.com
DOCKER_REPO=${DOCKER_REGISTRY}/${GITHUB_USERNAME}/${PROJECT}
COVERAGE_DIR=coverage

# Build the project
all: build

dep:
	GO111MODULE=on go get -v -d ./cmd/${PROJECT}/... ; \
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
	mkdir -p ./coverage; \
	go test ${PKG_LIST} -coverprofile ${COVERAGE_DIR}/coverage.cov; \
	go tool cover -html="${COVERAGE_DIR}/coverage.cov" -o ${COVERAGE_DIR}/coverage.html

linux: dep
	mkdir -p ./bin/; \
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${PROJECT}-linux-${GOARCH} ./cmd/${PROJECT}/...

rpi: dep
	mkdir -p ./bin/; \
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LDFLAGS} -o ./bin/${PROJECT}-linux-arm ./cmd/${PROJECT}/...

darwin: dep
	mkdir -p ./bin/; \
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${PROJECT}-darwin-${GOARCH} ./cmd/${PROJECT}/...

windows: dep
	mkdir -p ./bin/; \
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${PROJECT}-windows-${GOARCH}.exe ./cmd/${PROJECT}/...

docker_build:
	docker build -t ${DOCKER_REPO}:${VERSION} .
	docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:latest

docker_push:
	docker push ${DOCKER_REPO}:${VERSION}
	docker push ${DOCKER_REPO}:latest

docker: docker_build docker_push

build: linux darwin rpi windows

clean:
	-rm -rf ./bin/ ./${COVERAGE_DIR}/

.PHONY: linux darwin windows test fmt clean