.PHONY: test run clean all build buildwindows aiderupdate aiderinstalllinux aiderinstallwindows
all:
	test
clean:
	rm ./release/*
build:
	go build -o ./release/llm-tournament .
buildwindows:
	go build -o ./release/llm-tournament.exe .
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
