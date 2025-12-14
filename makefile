.PHONY: test run clean all build buildwindows setenv migrate aiderupdate aiderinstalllinux aiderinstallwindows update-coverage
all:
	test
clean:
	rm ./release/*
build:
	go build -o ./release/llm-tournament .
buildwindows:
	go build -o ./release/llm-tournament.exe .
setenv:
	go env -w CGO_ENABLED=1
migrate:
	go run main.go --migrate-to-sqlite
dedup:
	go run main.go --cleanup-duplicates
aiderinstalllinux:
	curl -LsSf https://aider.chat/install.sh | sh
aiderinstallwindows:
	powershell -ExecutionPolicy ByPass -c "irm https://aider.chat/install.ps1 | iex"
aiderupdate:
	aider --install-main-branch
test:
	CGO_ENABLED=1 go test -json ./... -race -cover 2>&1 | tdd-guard-go -project-root "C:/Users/lavantien/dev/llm-tournament"
test-verbose:
	CGO_ENABLED=1 go test ./... -v -race -cover
run:
	go run .
update-coverage:
	@CGO_ENABLED=1 go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep total
	@./scripts/update-badge.sh
