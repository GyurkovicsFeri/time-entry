package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gyurkovicsferi/time-tracker/lib/clockify"
	"github.com/gyurkovicsferi/time-tracker/lib/db"
	libStore "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2/query"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

var ClockifyCmd = &cli.Command{
	Name:        "clockify",
	Description: "Utilities for interacting with Clockify",
	Usage:       "clockify",
	Category:    "Clockify",
	Commands: []*cli.Command{
		{
			Name:           "config",
			Usage:          "config",
			Description:    "Clockify Configuration",
			DefaultCommand: "get",
			Commands: []*cli.Command{
				{
					Name:        "get",
					Usage:       "get",
					Description: "Get the Clockify configuration",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						store := clockify.NewClockifyStore(db.NewDB())
						config, err := store.GetClockifyConfig()
						if err != nil {
							return err
						}
						pterm.Println("API Key: " + config.APIKey)
						pterm.Println("Workspace ID: " + config.WorkspaceID)
						return nil
					},
				},
				{
					Name:        "set",
					Usage:       "set",
					Description: "Set the Clockify configuration",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						if cmd.Args().Len() != 2 {
							return fmt.Errorf("API key and workspace ID are required")
						}
						apiKey := cmd.Args().Get(0)
						workspaceId := cmd.Args().Get(1)
						store := clockify.NewClockifyStore(db.NewDB())
						store.InsertClockifyConfig(&clockify.ClockifyConfig{
							APIKey:      apiKey,
							WorkspaceID: workspaceId,
						})
						pterm.Println("Clockify API key and workspace ID saved")
						return nil
					},
				},
				{
					Name:        "delete",
					Usage:       "delete",
					Description: "Delete the Clockify configuration",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						store := clockify.NewClockifyStore(db.NewDB())
						store.DeleteClockifyConfig()
						pterm.Println("Clockify configuration deleted")
						return nil
					},
				},
			},
		},
		{
			Name:        "upload-last-week",
			Usage:       "upload-last-week",
			Description: "Upload time entries to Clockify for last week",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				now := time.Now()
				lastWeek := now.AddDate(0, 0, -7)
				startOfWeek := startOfWeek(lastWeek)
				endOfWeek := endOfWeek(lastWeek)

				return uploadTimeEntry(startOfWeek, endOfWeek)
			},
		},
		{
			Name:        "upload-today",
			Usage:       "upload-today",
			Description: "Upload time entries to Clockify for today",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				now := time.Now()
				startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				endOfDay := startOfDay.AddDate(0, 0, 1)
				return uploadTimeEntry(startOfDay, endOfDay)
			},
		},
	},
}

func startOfWeek(t time.Time) time.Time {
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	year, month, day := t.Year(), t.Month(), t.Day()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func endOfWeek(t time.Time) time.Time {
	for t.Weekday() != time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	year, month, day := t.Year(), t.Month(), t.Day()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location())
}

func uploadTimeEntry(start, end time.Time) error {
	newDb := db.NewDB()
	defer newDb.Close()

	store := libStore.NewStore(newDb)
	defer store.Close()

	timeEntries := store.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
		return q.Where(query.Field("start").GtEq(start).And(query.Field("start").LtEq(end)))
	})

	for _, timeEntry := range timeEntries {
		pterm.Println(timeEntry.ID)
	}

	clockifyStore := clockify.NewClockifyStore(newDb)
	clockifyConfig, err := clockifyStore.GetClockifyConfig()
	if err != nil {
		return err
	}

	for _, timeEntry := range timeEntries {
		clockifyTimeEntry, err := clockifyStore.GetClockifyTimeEntry(timeEntry)
		if err != nil {
			return err
		}
		api := clockify.NewClockifyAPI(clockifyConfig.APIKey, clockifyConfig.WorkspaceID)
		if clockifyTimeEntry != nil {
			// TODO: update the time entry
			if clockifyTimeEntry.Deleted {
				// TODO: delete the time entry
				pterm.Println("Deleting time entry: " + timeEntry.ID)
			} else {
				pterm.Println("Updating time entry: " + timeEntry.ID)
			}
		} else {
			pterm.Println("Creating time entry: " + timeEntry.Project + " " + timeEntry.Task)
			api.PostNewTimeEntry(timeEntry)
		}
	}

	return nil
}
