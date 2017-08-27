GOLINT_BIN         = $(shell go env GOPATH)/bin/golint
PKG_DIRS          := $(shell find . -type d | grep -v vendor)
FULL_PKGS         := $(sort $(foreach pkg, $(PKG_DIRS), $(subst ./, github.com/ohaiwalt/repo-gopher, $(pkg))))
BUILD_STAMP       := $(shell date -u '+%Y%m%d%H%M%S')
BUILD_HASH        := $(shell git rev-parse HEAD)
BUILD_TAG         ?= $(shell scripts/build_tag.sh)
DOCKER_IMAGE      ?= "ohaiwalt/repo-gopher:$(BUILD_TAG)"
LINK_VARS         := -X main.buildstamp=$(BUILD_STAMP) -X main.buildhash=$(BUILD_HASH)
LINK_VARS         += -X main.buildtag=$(BUILD_TAG)
BUILD_DIR          = _build
BINARY             = repo-gopher

ifdef FORCE
.PHONY: all test clean deps docker
else
.PHONY: all test clean deps docker
endif

all: test binary

deps:
	dep ensure

test: deps
	go test -cover

binary: clean-dev | $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LINK_VARS)" -o $(BUILD_DIR)/$(BINARY)

docker:
	docker build -t $(DOCKER_IMAGE) .

docker-push:
	docker push $(DOCKER_IMAGE)

run:
	docker run -v $(shell pwd)/examples/config.toml:/etc/repo-gopher/config.toml -e GITHUB_AUTH_TOKEN=$(GITHUB_AUTH_TOKEN) ohaiwalt/repo-gopher

clean: clean-dev
	rm -rf $(BUILD_DIR)
	find . -name "*.test" -type f | xargs rm -fv
	find . -name "*-test" -type f | xargs rm -fv

# Remove editor files (here, Emacs)
clean-dev:
	rm -f `find . -name "*flymake*.go"`

$(BUILD_DIR):
	mkdir -p $@
