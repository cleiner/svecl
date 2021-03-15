package main

import (
	"os"

	"github.com/cleiner/svecl/internal/cli"
)

func main() {
	result := cli.Run(os.Args[1:], sveclVersion)
	os.Exit(result)
}
