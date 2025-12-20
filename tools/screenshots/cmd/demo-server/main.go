package main

import (
	"log"
	"os"
)

var osExit = os.Exit

func main() {
	log.SetOutput(os.Stderr)
	osExit(run(os.Args[1:], defaultRunDeps()))
}
