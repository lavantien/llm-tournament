# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM := if exist release rmdir /s /q release
    MKDIR := if not exist release mkdir release
    CGO_PREFIX := set CGO_ENABLED=1 &&
    SHELL_EXT := .bat
    EXE_EXT := .exe
    GREP := findstr
    GRANT_PERMISSION := echo "Permission Granted"
    UPDATE_BADGE := powershell -ExecutionPolicy Bypass -File scripts\update-badge.ps1
else
    DETECTED_OS := $(shell uname -s)
    RM := rm -rf ./release/*
    MKDIR := mkdir -p ./release
    CGO_PREFIX := CGO_ENABLED=1
    SHELL_EXT := .sh
    EXE_EXT :=
    GREP := grep
    GRANT_PERMISSION := chmod +x ./scripts/update-badge.sh
    UPDATE_BADGE := ./scripts/update-badge.sh
endif

.PHONY: lint test testbrief cover check run clean all build buildwindows buildlinux setenv aiderupdate aiderinstalllinux aiderinstallwindows update-coverage update-coverage-table screenshots build-css watch-css clean-css verify-docs

all: lint test

clean:
	$(RM)
	rm data/tournament.db
	rm -f templates/output.css templates/output.css.map

build: build-css
ifeq ($(DETECTED_OS),Windows)
	$(MKDIR)
	$(CGO_PREFIX) go build -o ./release/llm-tournament.exe .
else
	$(MKDIR)
	$(CGO_PREFIX) go build -o ./release/llm-tournament .
endif

buildwindows:
	$(MKDIR)
	go build -o ./release/llm-tournament.exe .

buildlinux:
	$(MKDIR)
	CGO_ENABLED=1 go build -o ./release/llm-tournament .

setenv:
	go env -w CGO_ENABLED=1

aiderinstalllinux:
    curl -LsSf https://aider.chat/install.sh | sh

aiderinstallwindows:
	powershell -ExecutionPolicy ByPass -c "irm https://aider.chat/install.ps1 | iex"

aiderupdate:
	aider --install-main-branch

lint:
	golangci-lint run --no-config ./...

test:
	$(CGO_PREFIX) go test ./... -v -race -cover

testbrief:
	go test ./... -race -cover --coverprofile=coverage.out

cover:
	go tool cover -func coverage.out

check: setenv lint testbrief cover

run:
	$(CGO_PREFIX) go run .

update-coverage:
	@$(CGO_PREFIX) go test ./... -coverprofile=coverage.out
	@go tool cover -html coverage.out -o coverage.html
	@go tool cover -func coverage.out | $(GREP) total
	@$(GRANT_PERMISSION)
	@$(UPDATE_BADGE)

update-coverage-table:
	@$(CGO_PREFIX) go test ./... -coverprofile=coverage.out
	@go tool cover -html coverage.out -o coverage.html
	@go tool cover -func coverage.out | $(GREP) total
	@if [ "$(DETECTED_OS)" = "Windows" ]; then \
		powershell -ExecutionPolicy Bypass -File scripts/update-coverage-table.ps1; \
	else \
		chmod +x ./scripts/update-coverage-table.sh && ./scripts/update-coverage-table.sh; \
	fi

screenshots:
	npm run screenshots

verify-docs:
	@echo "Verifying documentation enforcement..."
	@$(CGO_PREFIX) go test -run TestDesignConceptAndPreview_ExistAndStructured -v
	@echo "Checking README.md coverage table format..."
	@python3 -c "import re, sys; f=open('README.md', encoding='utf-8'); c=f.read(); f.close(); m=re.search(r'###(?:\s+[\d.]+\s+)?Coverage.*?Package-level statement coverage from.*?\n\n\| Package \| Coverage \|\n\| --- \| ---: \|', c, re.DOTALL); sys.exit(0 if m else 1)" || (echo "ERROR: README.md Coverage section format is invalid" && exit 1)
	@echo "Documentation verification passed!"

build-css:
	npm run build:css

watch-css:
	npm run watch:css

clean-css:
	rm -f templates/output.css templates/output.css.map
