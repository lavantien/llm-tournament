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

.PHONY: lint test testbrief check run clean all build buildwindows buildlinux setenv aiderupdate aiderinstalllinux aiderinstallwindows update-coverage screenshots

all: lint test

clean:
	$(RM)
	rm data/tournament.db

build:
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
	go test ./... -race -cover

check: setenv lint testbrief

run:
	$(CGO_PREFIX) go run .

update-coverage:
	@$(CGO_PREFIX) go test ./... -coverprofile=coverage.out
	@go tool cover -html coverage.out -o coverage.html
	@go tool cover -func=coverage.out | $(GREP) total
	@$(GRANT_PERMISSION)
	@$(UPDATE_BADGE)

screenshots:
	npm run screenshots
