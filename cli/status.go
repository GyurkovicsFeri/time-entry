package main

import (
	"context"
	"fmt"
	"time"

	s "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

var StatusCmd = &cli.Command{
	Name:     "status",
	Aliases:  []string{"st"},
	Category: "reporting",
	Usage:    "Show the current time entry status",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "raw",
			Usage: "Show raw output. (Without colors)",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := s.NewStore()
		defer store.Close()

		current := store.GetCurrentTimeEntry()
		if current == nil {
			pterm.Println("No running time entry")
			return nil
		}

		printStatus(current, cmd.Bool("raw"))
		return nil
	},
}

func printStatus(current *s.CurrentTimeEntry, raw bool) {
	if raw {
		pterm.Println(current.Project, current.Task)
		return
	}

	title := pterm.NewStyle(pterm.Bold, pterm.FgBlue)
	info := pterm.NewStyle(pterm.FgGray)

	title.Print("Task: ")
	info.Print(current.Project, ", ", current.Task)
	pterm.Println()

	title.Print("Started at: ")
	info.Print(current.Start.Format(time.RFC850))
	pterm.Println()

	duration := time.Since(current.Start)
	formattedDuration := fmt.Sprintf("%dh %02dm", int(duration.Hours()), int(duration.Minutes())%60)
	title.Print("Duration: ")
	info.Print(formattedDuration)
}
