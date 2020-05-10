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

run: build
	./$(BINARY_NAME)

build:
	$(GOBUILD) -v -ldflags $(LDFLAGS) -o $(BINARY_NAME)

unit_test unit:
	GO_ENV=test $(GOTEST) ./... -parallel=10 -covermode=count #-v

integration_test int:
	GO_ENV=test $(GOTEST) -tags=integration ./integration_test

postman:
	newman run ".postman/Time Tracking API.postman_collection.json" --bail --ignore-redirects

vet:
	$(GOVET) ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

