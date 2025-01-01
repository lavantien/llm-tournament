.PHONY: updateaider test run clean all
all:
	test
clean:
updateaider:
	python -m pip install --upgrade git+https://github.com/paul-gauthier/aider.git
test:
	-go test ./... -v -race -cover > test_output.txt
run:
	go run .
