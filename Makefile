# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=knamespacer
COMMIT := $(shell git rev-parse HEAD)
VERSION := "dev"
ENVTEST_K8S_VERSION = 1.25.0

all: test build
build: tidy
	$(GOBUILD) -o $(BINARY_NAME) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -s -w" -v
lint:
	golangci-lint run
test:
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" \
	$(GOCMD) test -v --bench --benchmem -coverprofile coverage.txt -covermode=atomic ./...
	$(GOCMD) vet ./... 2> govet-report.out
	$(GOCMD) tool cover -html=coverage.txt -o cover-report.html
	printf "\nCoverage report available at cover-report.html\n\n"
e2e-test:
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" \
	$(GOCMD) test -v --bench --benchmem -coverprofile coverage.txt -covermode=atomic ./pkg/e2e
	$(GOCMD) vet ./... 2> govet-report.out
	$(GOCMD) tool cover -html=coverage.txt -o cover-report.html
	printf "\nCoverage report available at cover-report.html\n\n"
tidy:
	$(GOCMD) mod tidy
clean:
	$(GOCLEAN)
	$(GOCMD) fmt ./...
	rm -f $(BINARY_NAME)
	packr2 clean
	rm -rf e2e/results/*
	rm *-report*
	rm coverage.txt
	rm -f knamespacer-*.tgz
# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -ldflags "-X main.VERSION=$(VERSION)" -v
build-docker:
	docker build -t knamespacer:dev .
build-docker-test:
	docker build -t knamespacertest:dev -f testing.Dockerfile .
docker-test: build-docker-test
	docker run -it --rm -v "$(shell pwd):$(shell pwd)" -w "$(shell pwd)" knamespacertest:dev \
	/usr/bin/dumb-init make test
docker-e2e-test: build-docker-test
	docker run -it --rm -v "$(shell pwd):$(shell pwd)" -w "$(shell pwd)" knamespacertest:dev \
	/usr/bin/dumb-init make e2e-test
docker-test-shell: build-docker-test
	docker run -it --rm -v "$(shell pwd):$(shell pwd)" -w "$(shell pwd)" knamespacertest:dev \
	/bin/bash
docker-run: build-docker
	docker run -it --rm -v "${HOME}/.kube/:/root/.kube/" knamespacer:dev 
