PKG = github.com/tailor-platform/patterner
COMMIT = $(shell git rev-parse --short HEAD)

BUILD_LDFLAGS = "-s -w -X $(PKG)/version.Revision=$(COMMIT)"

default: test

ci: depsdev test

test:
	go test ./... -coverprofile=coverage.out -covermode=count -count=1

fulltest:
	env TEST_INTEGRATION=1 go test -v ./... -coverprofile=coverage.out -covermode=count -count=1

build:
	go build -ldflags=$(BUILD_LDFLAGS) -trimpath

install:
	go install -ldflags=$(BUILD_LDFLAGS) -trimpath

lint:
	golangci-lint run ./...

depsdev:
	go install github.com/Songmu/gocredits/cmd/gocredits@latest

prerelease_for_tagpr: depsdev
	go mod download
	gocredits -w .
	git add CHANGELOG.md CREDITS go.mod go.sum

.PHONY: default ci test fulltest build install lint depsdev prerelease_for_tagpr
