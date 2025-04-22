package main

import (
	"context"
	"fmt"
	"time"

	s "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2/query"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

var ListCmd = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "List all time entries",
	Flags: []cli.Flag{
		&cli.TimestampFlag{
			Name:  "from",
			Usage: "From date",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{"2006-01-02"},
			},
		},
		&cli.TimestampFlag{
			Name:  "to",
			Usage: "To date",
			Config: cli.TimestampConfig{
				Timezone: time.Local,
				Layouts:  []string{"2006-01-02"},
			},
		},
		&cli.BoolFlag{
			Name:  "today",
			Usage: "Show time entries for today",
		},
		&cli.BoolFlag{
			Name:  "yesterday",
			Usage: "Show time entries for yesterday",
		},
		&cli.BoolFlag{
			Name:  "id",
			Usage: "Show the id of the time entries",
		},
	},
	Category: "reporting",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := s.NewStore()
		defer store.Close()

		entries := store.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
			if HasFlag(cmd, "from") {
				from := cmd.Timestamp("from")
				return q.Where(query.Field("start").GtEq(from))
			}
			if HasFlag(cmd, "to") {
				to := cmd.Timestamp("to")
				return q.Where(query.Field("start").LtEq(to))
			}
			if cmd.Bool("today") {
				return q.Where(query.Field("start").
					GtEq(s.StartOfDay(time.Now())).
					And(query.Field("start").
						LtEq(s.EndOfDay(time.Now()))))
			}
			if cmd.Bool("yesterday") {
				return q.Where(query.Field("start").GtEq(s.StartOfDay(time.Now().Add(-24 * time.Hour)))).Where(query.Field("start").LtEq(s.EndOfDay(time.Now().Add(-24 * time.Hour))))
			}
			return q
		})

		showId := cmd.Bool("id")
		headers := []string{"Project", "Task", "Start", "End", "Duration"}

		if showId {
			headers = append([]string{"ID"}, headers...)
		}

		table := pterm.TableData{headers}
		for _, entry := range entries {
			duration := entry.End.Sub(entry.Start)

			row := []string{
				entry.Project,
				entry.Task,
				entry.Start.Format(time.Stamp),
				entry.End.Format(time.Stamp),
				fmt.Sprintf("%dh %02dm", int(duration.Hours()), int(duration.Minutes())%60),
			}

			if showId {
				row = append([]string{entry.ID}, row...)
			}

			table = append(table, row)
		}

		pterm.DefaultTable.
			WithHasHeader().
			WithData(table).
			WithAlternateRowStyle(pterm.NewStyle(pterm.BgBlack, pterm.BgDarkGray)).
			Render()

		return nil
	},
}
