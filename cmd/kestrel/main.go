package main

import (
	"os"

	"github.com/kestrelflow/kestrelflow/pkg/cli"
)

func main() {
	c := cli.NewCLI(os.Stdout, os.Stderr)
	code := c.Execute(os.Args[1:])
	os.Exit(code)
}
