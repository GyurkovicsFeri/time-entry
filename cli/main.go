package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
	te "gyurkovics.com/time-tracker/lib"
	s "gyurkovics.com/time-tracker/lib/store"
)

func main() {
	cmd := &cli.Command{
		Name:                  "time-entry",
		Usage:                 "Time entry CLI",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:      "start",
				Aliases:   []string{"s"},
				Usage:     "Start a time entry",
				ArgsUsage: "<project> <task>",
				Category:  "time-entry",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					te.Start(cmd.Args().First(), cmd.Args().Get(1), store)
					pterm.Println("Started time entry: ", cmd.Args().First(), cmd.Args().Get(1))

					return nil
				},
				ShellComplete: func(ctx context.Context, cmd *cli.Command) {
					if cmd.Args().Len() == 0 {
						fmt.Println("project1")
						fmt.Println("project2")
					} else if cmd.Args().Len() == 1 {
						fmt.Println("task1")
						fmt.Println("task2")
					}
				},
			},
			{
				Name:     "stop",
				Aliases:  []string{"e", "end"},
				Usage:    "Stop / end the current time entry",
				Category: "time-entry",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					current := store.GetCurrentTimeEntry()
					if current == nil {
						pterm.Println("No time entry to stop")
						return nil
					}

					te.Stop(store, current, time.Now())
					pterm.Println("Stopped time entry: ", current.Project, current.Task)

					return nil
				},
			},
			{
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
				},
				Category: "reporting",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					entries := store.GetTimeEntries()
					sort.Slice(entries, func(i, j int) bool {
						return entries[i].Start.Before(entries[j].Start)
					})

					if slices.Contains(cmd.FlagNames(), "from") && slices.Contains(cmd.FlagNames(), "to") {
						from := cmd.Timestamp("from")
						to := cmd.Timestamp("to")

						entries = filterEntries(entries, from, to)
					}

					if cmd.Bool("today") {
						entries = filterEntries(entries, s.StartOfDay(time.Now()), s.EndOfDay(time.Now()))
					}

					table := pterm.TableData{
						{"Project", "Task", "Start", "End", "Duration"},
					}

					for _, entry := range entries {
						duration := entry.End.Sub(entry.Start)
						table = append(table, []string{
							entry.Project,
							entry.Task,
							entry.Start.Format(time.Stamp),
							entry.End.Format(time.Stamp),
							fmt.Sprintf("%dh %02dm", int(duration.Hours()), int(duration.Minutes())%60),
						})
					}

					pterm.DefaultTable.
						WithHasHeader().
						WithData(table).
						WithAlternateRowStyle(pterm.NewStyle(pterm.BgBlack, pterm.BgDarkGray)).
						Render()

					return nil
				},
			},
			{
				Name:     "status",
				Aliases:  []string{"st"},
				Category: "reporting",
				Usage:    "Show the current time entry status",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					current := store.GetCurrentTimeEntry()
					if current == nil {
						pterm.Println("No time entry to stop")
						return nil
					}

					pterm.Println("Current time entry: ", current.Project, current.Task)
					pterm.Println("Started at: ", current.Start.Format(time.RFC3339))
					pterm.Println("Duration: ", time.Since(current.Start))

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func filterEntries(entries []*s.TimeEntry, from, to time.Time) []*s.TimeEntry {
	filtered := []*s.TimeEntry{}

	toUpper := to.Add(24 * time.Hour)

	for _, entry := range entries {
		if entry.Start.After(from) && entry.Start.Before(toUpper) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
