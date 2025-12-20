package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stderr)
	os.Exit(run(os.Args[1:], defaultRunDeps()))
}
