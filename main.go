package main

import (
	"os"

	"github.com/user/speed-test-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
