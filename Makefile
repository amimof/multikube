# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

BINARY=multikube
GOARCH=amd64
VERSION=1.0.0
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
GITHUB_USERNAME=amimof
BUILD_DIR=${GOPATH}/src/gitlab.com/${GITHUB_USERNAME}/${BINARY}

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

# Build the project
all: test fmt 

test:
	cd ${BUILD_DIR}; \
	go test ; \
	cd - >/dev/null

fmt:
	cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

generate-api: 
	cd ${BUILD_DIR}
	swagger validate api/v1/swagger.yml
	swagger generate server -A ${BINARY} -s server -a restapi -t api/v1/ -f api/v1/swagger.yml

# clean:
# 	-rm -f ${BINARY}-*

.PHONY: linux darwin windows test fmt clean