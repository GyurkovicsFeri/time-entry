package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:                  "time-entry",
		Usage:                 "Time entry CLI",
		EnableShellCompletion: true,
		Suggest:               true,
		Commands: []*cli.Command{
			StartCmd,
			StopCmd,
			ListCmd,
			StatusCmd,
			EditCmd,
			DeleteCmd,
			ReportCmd,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
