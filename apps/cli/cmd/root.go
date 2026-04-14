package cmd

import (
	"context"

	"github.com/urfave/cli/v3"
)

// global verbose flag - accessed by commands
var Verbose bool

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "interlock",
		Usage: "setup an interlock project",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "enable verbose output",
				Action: func(ctx context.Context, cmd *cli.Command, v bool) error {
					Verbose = v
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			NewInitCommand(),
			NewProjectCommandGroup(),
		},
	}
}

func NewProjectCommandGroup() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "project-related commands",
		Commands: []*cli.Command{
			NewBundleCommand(),
			NewRunCommand(),
			NewDevCommand(),
		},
	}
}
