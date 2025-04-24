package main

import (
	"context"
	"slices"
	"time"

	timeentry "github.com/gyurkovicsferi/time-tracker/lib"
	"github.com/gyurkovicsferi/time-tracker/lib/db"
	s "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

var StopCmd = &cli.Command{
	Name:    "stop",
	Aliases: []string{"e", "end"},
	Usage:   "Stop / end the current time entry",
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name:  "end",
			Usage: "End the time entry at a specific time",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{"2006-01-02 15:04:05"},
			},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		db := db.NewDB()
		defer db.Close()

		store := s.NewStore(db)
		defer store.Close()

		end := time.Now()
		if slices.Contains(cmd.FlagNames(), "end") {
			end = cmd.Timestamp("end").Local()
		}

		current := store.GetCurrentTimeEntry()
		if current == nil {
			pterm.Println("No time entry to stop")
			return nil
		}

		timeentry.Stop(store, current, end)
		pterm.Println("Stopped time entry: ", current.Project, current.Task)

		return nil
	},
}
