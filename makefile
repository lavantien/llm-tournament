.PHONY: updateaider test run clean all
all:
	test
clean:
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
