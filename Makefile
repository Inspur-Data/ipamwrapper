SHELL = /usr/bin/env bash -o pipefail
# Current ipam version
VERSION = 0.0.1
RELEASE_TAG = $(shell cat VERSION)
DATE = $(shell date +"%Y-%m-%d_%H:%M:%S")
REGISTRY = inspurwyd
IMAGENAME =  k8-ipam
# image tag
IPAM_IMG = ${IMAGENAME}:$(VERSION)
GOLDFLAGS = "-w -s -extldflags '-z now' -X github.com/Inspur-Data/${IMAGENAME}/versions.COMMIT=$(COMMIT) -X github.com/Inspur-Data/${IMAGENAME}/versions.VERSION=$(RELEASE_TAG) -X github.com/Inspur-Data/${IMAGENAME}/versions.BUILDDATE=$(DATE)"

GLDFLAGS+="-X ${REPO}/pkg/version.Raw=${VERSION_OVERRIDE}"

.PHONY: build-bin
build-bin:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildmode=pie -o bin/${IMAGENAME} -ldflags ${GLDFLAGS}  -v ./cmd/${IMAGENAME}-ds

.PHONY: build-k8-ipam
build-k8-ipam: build-bin
	docker build -t $(REGISTRY)/${IMAGENAME}:$(RELEASE_TAG) --build-arg VERSION=$(RELEASE_TAG) .
