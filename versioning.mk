GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_SHA := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif

BINARY_VERSION ?= ${GIT_TAG}-${GIT_SHA}

LDFLAGS += -X github.com/helm/prow/pkg/version.Version=${GIT_TAG}
LDFLAGS += -X github.com/helm/prow/pkg/version.GitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/helm/prow/pkg/version.GitTreeState=${GIT_DIRTY}
