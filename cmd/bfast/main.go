package main

import (
	"context"
	"os"

	"github.com/arrno/bfast/internal/cli"
)

func main() {
	ctx := context.Background()
	code := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}
