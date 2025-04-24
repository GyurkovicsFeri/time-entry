package main

import (
	"context"
	"fmt"

	"github.com/gyurkovicsferi/time-tracker/lib/clockify"
	"github.com/gyurkovicsferi/time-tracker/lib/db"
	store "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2/query"
	"github.com/urfave/cli/v3"
)

var DeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a time entry",
	ArgsUsage: "[id]",
	Category:  "time-entry",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "last",
			Usage: "Delete the last time entry",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		db := db.NewDB()

		store := store.NewStore(db)
		defer store.Close()

		if cmd.Bool("last") {
			store.DeleteTimeEntry(store.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
				return q.Limit(1).Sort(query.SortOption{
					Field:     "start",
					Direction: -1,
				})
			})[0].ID)
		} else {
			if cmd.Args().First() == "" {
				return fmt.Errorf("id is required")
			}
			store.DeleteTimeEntry(cmd.Args().First())
		}

		clockifyStore := clockify.NewClockifyStore(db)
		clockifyStore.MakeClockifyTimeEntryDeleted(cmd.Args().First())

		return nil
	},
}
