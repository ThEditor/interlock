package cmd

import "github.com/urfave/cli/v3"

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "interlock",
		Usage: "setup an interlock project",
		Commands: []*cli.Command{
			NewInitCommand(),
		},
	}
}
