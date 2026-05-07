ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=SET
    WHICH=where.exe
    DEL=del
    NUL=nul
else
    SET=export
    WHICH=which
    DEL=rm
    NUL=/dev/null
endif

ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL) || echo go)
endif

NAME:=$(notdir $(CURDIR))
VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
GOOPT:=-ldflags "-s -w -X main.version=$(VERSION)"
EXE:=$(shell $(GO) env GOEXE)

build:
	$(GO) fmt ./...
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)

define _build
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT) "$1"

endef

all:
	go fmt ./...
	$(foreach i,$(wildcard ./cmd/*),$(call _build,$(i)))

_dist:
	$(SET) "CGO_ENABLED=0" && go build $(GOOPT)
	zip $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=linux"   && $(SET) "GOARCH=amd64" && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist

release:
	$(GO) run github.com/hymkor/latest-notes@latest | gh release create -d --notes-file - -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

manifest:
	$(GO) run github.com/hymkor/make-scoop-manifest@latest -all *-windows-*.zip > $(NAME).json
