GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

BINARY_NAME=timetrack

all: run

run: build test
	./$(BINARY_NAME)

build:
	$(GOBUILD) -v -o $(BINARY_NAME)

test:
	GO_ENV=test $(GOTEST) ./... -v -parallel=10 -covermode=count

vet:
	$(GOVET) ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

