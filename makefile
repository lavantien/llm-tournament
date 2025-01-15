.PHONY: updateaider test run clean all build
all:
	test
clean:
	rm ./release/*
build:
	go build -o ./release/llm-tournament-v1.0 .
aiderinstalllinux:
	curl -LsSf https://aider.chat/install.sh | sh
aiderinstallwindows:
	powershell -ExecutionPolicy ByPass -c "irm https://aider.chat/install.ps1 | iex"
aiderupdate:
	aider --install-main-branch
test:
	-go test ./... -v -race -cover > test_output.txt
run:
	go run .
