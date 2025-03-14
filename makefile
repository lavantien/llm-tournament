.PHONY: test run clean all build buildwindows setenv migrate aiderupdate aiderinstalllinux aiderinstallwindows
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
	-go test ./... -v -race -cover
run:
	go run .
