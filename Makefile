# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

BINARY=multikube
GOARCH=amd64
VERSION=1.0.0-alpha.6
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GOVERSION=$(shell go version | awk -F\go '{print $$3}' | awk '{print $$1}')
GITHUB_USERNAME=amimof
BUILD_DIR=${GOPATH}/src/gitlab.com/${GITHUB_USERNAME}/${BINARY}
PKG_LIST=$$(go list ./... | grep -v /vendor/)
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -X main.GOVERSION=${GOVERSION}"

# Build the project
all: build

test:
	cd ${BUILD_DIR}; \
	go test ; \
	cd - >/dev/null

fmt:
	cd ${BUILD_DIR}; \
	go fmt ${PKG_LIST} ; \
	cd - >/dev/null

dep:
	go get -v -d ./cmd/multikube/... ;

linux: dep
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BUILD_DIR}/out/${BINARY}-linux-${GOARCH} cmd/multikube/main.go

rpi: dep
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LDFLAGS} -o ${BUILD_DIR}/out/${BINARY}-linux-arm cmd/multikube/main.go

darwin: dep
	CGO_ENABLED=0 GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BUILD_DIR}/out/${BINARY}-darwin-${GOARCH} cmd/multikube/main.go

windows: dep
	CGO_ENABLED=0 GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BUILD_DIR}/out/${BINARY}-windows-${GOARCH}.exe cmd/multikube/main.go

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
	-rm -rf ${BUILD_DIR}/out/

.PHONY: linux darwin windows test fmt clean