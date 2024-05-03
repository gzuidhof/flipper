package cmd

import (
	"context"

	"github.com/gzuidhof/flipper/buildinfo"
	"github.com/gzuidhof/flipper/entry"
	"github.com/urfave/cli/v3"
)

// CLI is the entrypoint for the flipper binary CLI.
func CLI(ctx context.Context, args []string) error {
	c := &cli.Command{
		Name:                  "flipper",
		Usage:                 "A tool that watches cloud resources and re-points re-assignable IPs to healthy servers.",
		Version:               buildinfo.FullVersion(),
		EnableShellCompletion: true,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Path to the configuration file",
				Aliases: []string{"c"},
				Value:   "flipper.yaml",
			},
		},

		Commands: []*cli.Command{
			{
				Name:   "monitor",
				Usage:  "Start monitoring the resources",
				Action: func(ctx context.Context, c *cli.Command) error { return entry.Monitor(ctx, c.String("config")) },
			},
		},
	}

	//nolint:wrapcheck // No point in wrapping the error here.
	return c.Run(ctx, args)
}
