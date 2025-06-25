package main

import (
	"os"

	"github.com/thinkgos/proc-extra/cmd/enumgen/command"
)

var cmd = command.NewRootCmd()

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
