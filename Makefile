PKG = github.com/jarnfast/signalman
SHELL = bash

GO = GO111MODULE=on go

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo main)

#LDFLAGS = -w -X $(PKG)/pkg.Version=$(VERSION) -X $(PKG)/pkg.Commit=$(COMMIT)
#optimize flags -w -s
#LDFLAGS = -w -X main.version=$(VERSION) -X main.build=$(COMMIT)
LDFLAGS = -w -X jarnfast/signalman/pkg.Version=$(VERSION) -X jarnfast/signalman/pkg.Build=$(COMMIT)
XBUILD = CGO_ENABLED=0 $(GO) build -a -tags netgo -ldflags '$(LDFLAGS)'

PLATFORM ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

BINDIR = bin

MIXIN = signalman

ifeq ($(PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

.PHONY: build
build: 
	mkdir -p $(BINDIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(MIXIN)$(FILE_EXT) ./cmd/$(MIXIN)

.PHONY: run
run: build
	./$(BINDIR)

xbuild-all:	
	$(MAKE) $(MAKE_OPTS) PLATFORM=windows ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=darwin ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=arm64 xbuild;

xbuild: $(BINDIR)/$(VERSION)/$(MIXIN)-$(PLATFORM)-$(ARCH)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(MIXIN)-$(PLATFORM)-$(ARCH)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(PLATFORM) GOARCH=$(ARCH) $(XBUILD) -o $@ ./cmd/$(MIXIN)

test: test-unit
	$(BINDIR)/$(MIXIN)$(FILE_EXT) version

test-unit: build
	$(GO) test ./...

test-integration: xbuild
	# Test against the cross-built client binary that we will publish
	cp $(BINDIR)/$(VERSION)/$(MIXIN)-$(PLATFORM)-$(ARCH)$(FILE_EXT) $(BINDIR)/$(MIXIN)$(FILE_EXT)
	$(GO) test -tags=integration ./tests/...

clean:
	-rm -fr bin/
