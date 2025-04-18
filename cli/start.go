package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	timeentry "github.com/gyurkovicsferi/time-tracker/lib"
	store "github.com/gyurkovicsferi/time-tracker/lib/store"
)

var StartCmd = &cli.Command{
	Name:      "start",
	Aliases:   []string{"s"},
	Usage:     "Start a time entry",
	ArgsUsage: "<project> <task>",
	Category:  "time-entry",
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name:  "from",
			Usage: "Start the time entry at a specific time",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{"2006-01-02 15:04:05"},
			},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.Args().Len() != 2 {
			return fmt.Errorf("project and task are required")
		}

		store := store.NewStore()
		defer store.Close()

		from := time.Now()
		if HasFlag(cmd, "from") {
			from = cmd.Timestamp("from")
		}

		timeentry.NewCurrentTimeEntry(store, cmd.Args().First(), cmd.Args().Get(1), from)

		pterm.NewStyle(pterm.FgGreen).Println("Started time entry: ", cmd.Args().First(), cmd.Args().Get(1))
		return nil
	},
	ShellComplete: func(ctx context.Context, cmd *cli.Command) {
		if cmd.Args().Len() == 0 {
			store := store.NewStore()
			defer store.Close()

			projects := store.GetProjects()
			for _, project := range projects {
				fmt.Println(project)
			}
		} else if cmd.Args().Len() == 1 {
			store := store.NewStore()
			defer store.Close()

			tasks := store.GetTasks(cmd.Args().First())
			for _, task := range tasks {
				fmt.Println(task)
			}
		}
	},
}
