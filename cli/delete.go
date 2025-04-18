package main

import (
	"context"
	"fmt"

	store "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2/query"
	"github.com/urfave/cli/v3"
)

var DeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "Delete a time entry",
	ArgsUsage: "[id]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "last",
			Usage: "Delete the last time entry",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		db := store.NewStore()
		defer db.Close()

		if cmd.Bool("last") {
			db.DeleteTimeEntry(db.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
				return q.Limit(1).Sort(query.SortOption{
					Field:     "start",
					Direction: -1,
				})
			})[0].ID)
		} else {
			if cmd.Args().First() == "" {
				return fmt.Errorf("id is required")
			}
			db.DeleteTimeEntry(cmd.Args().First())
		}

		return nil
	},
}
