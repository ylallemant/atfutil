package main

import (
	"os"

	"atfutil/pkg/cli/root"
)

func main() {
	if err := root.Command().Execute(); err != nil {
		os.Exit(1)
	}
}
