package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"time"

	te "github.com/gyurkovicsferi/time-tracker/lib"
	s "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
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
					store := s.NewStore()
					defer store.Close()

					from := time.Now()
					if slices.Contains(cmd.FlagNames(), "from") {
						from = cmd.Timestamp("from")
					}

					te.NewCurrentTimeEntry(store, cmd.Args().First(), cmd.Args().Get(1), from)

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
					store := s.NewStore()
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

					te.Stop(store, current, end)
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
					&cli.BoolFlag{
						Name:  "yesterday",
						Usage: "Show time entries for yesterday",
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

					if cmd.Bool("yesterday") {
						pterm.Info.Println("Showing time entries for yesterday")
						yesterday := time.Now().Add(-24 * time.Hour)
						pterm.Info.Println("From: ", s.StartOfDay(yesterday))
						pterm.Info.Println("To: ", s.EndOfDay(yesterday))
						entries = filterEntries(entries, s.StartOfDay(time.Now().Add(-24*time.Hour)), s.EndOfDay(time.Now().Add(-24*time.Hour)))
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
					pterm.Println()

					return nil
				},
			},
			{
				Name:     "report",
				Aliases:  []string{"r"},
				Usage:    "Generate a report",
				Category: "reporting",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					entries := store.GetTimeEntries()

					sort.Slice(entries, func(i, j int) bool {
						return entries[i].Start.Before(entries[j].Start)
					})

					for _, entry := range entries {
						pterm.Println(entry.Project, entry.Task, entry.Start, entry.End)
					}

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

	for _, entry := range entries {
		if entry.Start.After(from) && entry.Start.Before(to) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
