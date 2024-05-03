// Package main is the entrypoint for the flipper binary.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gzuidhof/flipper/cmd"
)

func run(ctx context.Context, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err := cmd.CLI(ctx, args)
	if err != nil {
		//nolint:wrapcheck // No point in wrapping the error here.
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
