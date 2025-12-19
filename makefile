# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM := if exist release rmdir /s /q release
    MKDIR := if not exist release mkdir release
    CGO_PREFIX := set CGO_ENABLED=1 &&
    SHELL_EXT := .bat
    EXE_EXT := .exe
    GREP := findstr
    UPDATE_BADGE := powershell -ExecutionPolicy Bypass -File scripts\update-badge.ps1
else
    DETECTED_OS := $(shell uname -s)
    RM := rm -rf ./release/*
    MKDIR := mkdir -p ./release
    CGO_PREFIX := CGO_ENABLED=1
    SHELL_EXT := .sh
    EXE_EXT :=
    GREP := grep
    UPDATE_BADGE := ./scripts/update-badge.sh
endif

.PHONY: test run clean all build buildwindows buildlinux setenv aiderupdate aiderinstalllinux aiderinstallwindows update-coverage

all: test

clean:
	$(RM)

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

test:
	$(CGO_PREFIX) go test -json ./... -race -cover 2>&1 | tdd-guard-go -project-root "C:/Users/lavantien/dev/llm-tournament"

test-verbose:
	$(CGO_PREFIX) go test ./... -v -race -cover

run:
	$(CGO_PREFIX) go run .

update-coverage:
	@$(CGO_PREFIX) go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | $(GREP) total
	@$(UPDATE_BADGE)
