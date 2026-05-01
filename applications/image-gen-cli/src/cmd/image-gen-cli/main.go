package main

import (
	"os"

	"github.com/ben-wangz/bot-cli/applications/image-gen-cli/src/internal/cli"
)

func main() {
	exitCode := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
