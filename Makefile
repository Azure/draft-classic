DOCKER_REGISTRY ?= quay.io
IMAGE_PREFIX    ?= deis
IMAGE_TAG       ?= canary
SHORT_NAME      ?= prowd
TARGETS         = darwin/amd64 linux/amd64 linux/386 linux/arm windows/amd64
DIST_DIRS       = find * -type d -exec
APP             = prow

# go option
GO        ?= go
PKG       := $(shell glide novendor)
TAGS      := kqueue
TESTS     := .
TESTFLAGS :=
LDFLAGS   :=
GOFLAGS   :=
BINDIR    := $(CURDIR)/bin
BINARIES  := prow

# Required for globs to work correctly
SHELL=/bin/bash

.PHONY: all
all: build

.PHONY: build
build:
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/prow/cmd/...

# usage: make clean build-cross dist APP=prow|prowd VERSION=v2.0.0-alpha.3
.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross:
	CGO_ENABLED=0 gox -output="_dist/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/prow/cmd/$(APP)

.PHONY: dist
dist:
	( \
		cd _dist && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf prow-${VERSION}-{}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r prow-${VERSION}-{}.zip {} \; && \
		mv $(APP)-${VERSION}-*.* .. \
	)

.PHONY: checksum
checksum:
	for f in _dist/*.{gz,zip} ; do \
		shasum -a 256 "$${f}"  | awk '{print $$1}' > "$${f}.sha256" ; \
	done

.PHONY: check-docker
check-docker:
	@if [ -z $$(which docker) ]; then \
	  echo "Missing \`docker\` client which is required for development"; \
	  exit 2; \
	fi

.PHONY: check-helm
check-helm:
	@if [ -z $$(which helm) ]; then \
	  echo "Missing \`helm\` client which is required for development"; \
	  exit 2; \
	fi

.PHONY: docker-binary
docker-binary: BINDIR = ./rootfs/bin
docker-binary: GOFLAGS += -a -installsuffix cgo
docker-binary:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o $(BINDIR)/prowd $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' github.com/deis/prow/cmd/prowd

.PHONY: docker-build
docker-build: check-docker docker-binary compress-binary
	docker build --rm -t ${IMAGE} .
	docker tag ${IMAGE} ${MUTABLE_IMAGE}

.PHONY: compress-binary
compress-binary: BINDIR = ./rootfs/bin
compress-binary:
	@if [ -z $$(which upx) ]; then \
	  echo "Missing \`upx\` tool to compress binaries"; \
	else \
	  upx --quiet ${BINDIR}/prowd; \
	fi

.PHONY: serve
serve: check-helm
	helm install chart/ --name ${APP} --namespace kube-system \
		--set image.name=${SHORT_NAME},image.org=${IMAGE_PREFIX},image.registry=${DOCKER_REGISTRY},image.tag=${IMAGE_TAG}

.PHONY: unserve
unserve: check-helm
	-helm delete --purge ${APP}

.PHONY: clean
clean:
	-rm bin/*
	-rm rootfs/bin/*

.PHONY: test
test: TESTFLAGS += -race -v
test: test-unit

.PHONY: test-unit
test-unit:
	$(GO) test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

.PHONY: test-e2e
test-e2e:
	./tests/e2e.sh

HAS_GLIDE := $(shell command -v glide;)
HAS_GOX := $(shell command -v gox;)
HAS_GIT := $(shell command -v git;)

.PHONY: bootstrap
bootstrap:
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
ifndef HAS_GOX
	go get -u github.com/mitchellh/gox
endif
ifndef HAS_GIT
	$(error You must install git)
endif
	glide install

include versioning.mk
