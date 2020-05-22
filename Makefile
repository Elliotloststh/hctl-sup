GO ?= go
PROJECT := github.com/elliotloststh/hctl-sup
BINDIR := /usr/local/bin
ifeq ($(GOPATH),)
export GOPATH := $(CURDIR)/_output
unexport GOBIN
endif
GOBINDIR := $(word 1,$(subst :, ,$(GOPATH)))
PATH := $(GOBINDIR)/bin:$(PATH)
GOPKGDIR := $(GOPATH)/src/$(PROJECT)
GOPKGBASEDIR := $(shell dirname "$(GOPKGDIR)")

VERSION := $(shell git describe --tags --dirty --always)
VERSION := $(VERSION:v%=%)
GO_LDFLAGS := -X $(PROJECT)/pkg/version.Version=$(VERSION)

all: binaries

help:
	@echo "Usage: make <target>"
	@echo
	@echo " * 'install' - Install binaries to system locations."
	@echo " * 'binaries' - Build hctl-sup."
	@echo " * 'clean' - Clean artifacts."

check-gopath:
ifeq ("$(wildcard $(GOPKGDIR))","")
	mkdir -p "$(GOPKGBASEDIR)"
	ln -s "$(CURDIR)" "$(GOPKGBASEDIR)/hctl-sup"
endif
ifndef GOPATH
	$(error GOPATH is not set)
endif

hctl-sup: check-gopath
		CGO_ENABLED=0 $(GO) install \
		-ldflags '$(GO_LDFLAGS)' \
		$(PROJECT)/cmd/hctl-sup

clean:
	find . -name \*~ -delete
	find . -name \#\* -delete

binaries: hctl-sup

install-hctl-sup: check-gopath
	install -D -m 755 $(GOBINDIR)/bin/hctl-sup $(BINDIR)/hctl-sup

install: install-hctl-sup


uninstall-hctl-sup:
		rm -f $(BINDIR)/hctl-sup

uninstall: uninstall-hctl-sup

.PHONY: \
	help \
	check-gopath \
	hctl-sup \
	clean \
	binaries \
	install \
	install-hctl-sup \
	uninstall \
	uninstall-hctl-sup \
