package main

import (
	"os"

	"github.com/garaemon/paperpile/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
