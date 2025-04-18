package main

import (
	"slices"

	"github.com/urfave/cli/v3"
)

func HasFlag(cmd *cli.Command, flag string) bool {
	return slices.Contains(cmd.FlagNames(), flag)
}
