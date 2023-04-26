PKG = github.com/jarnfast/signalman
SHELL = bash

GO = GO111MODULE=on go

COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags 2> /dev/null || echo main)

GOPATH ?= $(shell go env GOPATH)

LDFLAGS = -s -w -X jarnfast/signalman/pkg.Version=$(VERSION) -X jarnfast/signalman/pkg.Build=$(COMMIT)
XBUILD = CGO_ENABLED=0 $(GO) build -a -gcflags "all=-trimpath=$(GOPATH)" -ldflags '$(LDFLAGS)'

PLATFORM ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

BINDIR = bin

BINNAME = signalman

ifeq ($(PLATFORM),windows)
FILE_EXT=.exe
else
FILE_EXT=
endif

ifeq ($(ARCH),arm)
ARMVERSION=-v$(GOARM)
DOCKERARMVERSION=/$(GOARM)
else
ARMVERSION=
DOCKERARMVERSION=
endif

.PHONY: build
build:
	mkdir -p $(BINDIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME)$(FILE_EXT) ./cmd/$(BINNAME)

.PHONY: run
run: build
	./$(BINDIR)

.PHONY: xbuild-freebsd
xbuild-freebsd:
	rm bin/main/signalman-freebsd-amd64
	$(MAKE) $(MAKE_OPTS) PLATFORM=freebsd ARCH=amd64 xbuild;

.PHONY: xbuild-all
xbuild-all:
	$(MAKE) $(MAKE_OPTS) PLATFORM=windows ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=darwin ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=amd64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=arm64 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=arm GOARM=6 xbuild;
	$(MAKE) $(MAKE_OPTS) PLATFORM=linux ARCH=arm GOARM=7 xbuild;

.PHONY: xbuild
xbuild: $(BINDIR)/$(VERSION)/$(BINNAME)-$(PLATFORM)-$(ARCH)$(ARMVERSION)$(FILE_EXT)
$(BINDIR)/$(VERSION)/$(BINNAME)-$(PLATFORM)-$(ARCH)$(ARMVERSION)$(FILE_EXT):
	mkdir -p $(dir $@)
	GOOS=$(PLATFORM) GOARCH=$(ARCH) GOARM=$(GOARM) $(XBUILD) -o $@ ./cmd/$(BINNAME)

.PHONY: test
test: test-unit
	$(BINDIR)/$(BINNAME)$(FILE_EXT) version

.PHONY: test-unit
test-unit: build
	$(GO) test ./...

.PHONY: test-integration
test-integration: xbuild
	cp $(BINDIR)/$(VERSION)/$(BINNAME)-$(PLATFORM)-$(ARCH)$(FILE_EXT) $(BINDIR)/$(BINNAME)$(FILE_EXT)
	$(GO) test -tags=integration ./tests/...

.PHONY: clean
clean:
	rm -fr bin/

.PHONY: docker-buildx
docker-buildx:
	docker buildx build --build-arg VERSION=$(VERSION) -t signalman-init:latest -f Dockerfile.signalman-init --platform $(PLATFORM)/$(ARCH)$(DOCKERARMVERSION) .

.PHONY: generate-checksums
generate-checksums:
	cd bin/$(VERSION) && find . -type f -exec sh -c "sha256sum -b {} > {}.sha256sum" \;
