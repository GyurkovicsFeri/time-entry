package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
				Name:     "list",
				Aliases:  []string{"l"},
				Usage:    "List all time entries",
				Category: "reporting",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					entries := store.GetTimeEntries()

					table := pterm.TableData{
						{"ID", "Project", "Task", "Start", "End"},
					}

					for _, entry := range entries {
						table = append(table, []string{
							entry.ID,
							entry.Project,
							entry.Task,
							entry.Start.Format(time.RFC3339),
							entry.End.Format(time.RFC3339),
						})
					}

					pterm.DefaultTable.WithHasHeader().WithData(table).Render()

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
			{
				Name:     "today",
				Aliases:  []string{"t"},
				Category: "reporting",
				Usage:    "Show the time entries for today",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					store := s.NewStore()
					defer store.Close()

					entries := store.GetTimeEntries()

					pterm.Println("Time entries for today:")
					table := [][]string{
						{"ID", "Project", "Task", "Start", "End"},
					}

					for _, entry := range entries {
						table = append(table, []string{
							entry.ID,
							entry.Project,
							entry.Task,
							entry.Start.Format(time.RFC3339),
							entry.End.Format(time.RFC3339),
						})
					}

					pterm.DefaultTable.WithHasHeader().WithData(table).Render()

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
