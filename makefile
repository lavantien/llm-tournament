.PHONY: updateaider test run clean all
all:
	test
clean:
updateaider:
	curl -LsSf https://aider.chat/install.sh | sh
test:
	-go test ./... -v -race -cover > test_output.txt
run:
	go run .
