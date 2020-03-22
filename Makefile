GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

RELEASE?=1.0.0
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
PROJECT_ROOT=github.com/bryanmorgan/time-tracking-api
LDFLAGS="-s -w -X ${PROJECT_ROOT}/version.Release=${RELEASE} -X ${PROJECT_ROOT}/version.Commit=${COMMIT} -X ${PROJECT_ROOT}/version.BuildTime=${BUILD_TIME}"

BINARY_NAME=timetrack

all: run

run: build test
	./$(BINARY_NAME)

build:
	$(GOBUILD) -v -ldflags $(LDFLAGS) -o $(BINARY_NAME)

test:
	GO_ENV=test $(GOTEST) ./... -v -parallel=10 -covermode=count

vet:
	$(GOVET) ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

