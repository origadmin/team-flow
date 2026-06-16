# team-flow Makefile
# v3 flow tool — single binary, project-local build

BINARY  := flow
PKG     := ./cmd/flow
DESTDIR ?= $(GOPATH)/bin

# Use .exe suffix on Windows, no suffix on Unix
ifeq ($(OS),Windows_NT)
	EXE := .exe
else
	EXE :=
endif

PROJECT_BIN := ./scripts/$(BINARY)$(EXE)

.PHONY: build install uninstall clean test version

build:
	go build -o $(PROJECT_BIN) $(PKG)

install: build
ifeq ($(OS),Windows_NT)
	copy /Y $(PROJECT_BIN) $(DESTDIR)\$(BINARY)$(EXE)
else
	install -m 0755 $(PROJECT_BIN) $(DESTDIR)/$(BINARY)
endif
	@echo "Installed to $(DESTDIR)/$(BINARY)$(EXE)"
	@echo "Verify: flow --version"

uninstall:
	rm -f $(DESTDIR)/$(BINARY)$(EXE)

test:
	go test ./internal/...

version:
	go version
	@$(PROJECT_BIN) --version 2>/dev/null || echo "(binary not built yet — run 'make build')"

clean:
	rm -f $(PROJECT_BIN)
