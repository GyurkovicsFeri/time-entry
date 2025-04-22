package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	s "github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2/query"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

var ReportCmd = &cli.Command{
	Name:     "report",
	Aliases:  []string{"r"},
	Usage:    "Generate a report",
	Category: "reporting",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "this-week",
			Aliases: []string{"tw"},
			Usage:   "Show report for this week",
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "last-week",
			Aliases: []string{"lw"},
			Usage:   "Show report for last week",
		},
		&cli.BoolFlag{
			Name:    "today",
			Aliases: []string{"td"},
			Usage:   "Show report for today",
		},
		&cli.StringFlag{
			Name:    "project",
			Aliases: []string{"p"},
			Usage:   "Filter report by project",
		},
		&cli.StringFlag{
			Name:    "task",
			Aliases: []string{"t"},
			Usage:   "Filter report by task",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		store := s.NewStore()
		defer store.Close()

		// Determine time period based on flags
		var startDate, endDate time.Time
		var periodStr string

		if cmd.Bool("last-week") {
			// Last week
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startDate = now.AddDate(0, 0, -weekday-7+1).Truncate(24 * time.Hour)
			endDate = startDate.AddDate(0, 0, 7).Add(-time.Nanosecond)
			periodStr = "Last Week"
		} else if cmd.Bool("today") {
			// Today
			startDate = s.StartOfDay(time.Now())
			endDate = s.EndOfDay(time.Now())
			periodStr = "Today"
		} else {
			// This week (default)
			now := time.Now()
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startDate = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			endDate = now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
			periodStr = "This Week"
		}

		// Get entries for the selected time period and apply optional filters
		entries := store.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
			// Apply time range filter
			q = q.Where(query.Field("start").GtEq(startDate).And(query.Field("start").LtEq(endDate)))

			// Apply project filter if provided
			if project := cmd.String("project"); project != "" {
				q = q.Where(query.Field("project").Eq(project))
				periodStr += fmt.Sprintf(" (Project: %s)", project)
			}

			// Apply task filter if provided
			if task := cmd.String("task"); task != "" {
				q = q.Where(query.Field("task").Eq(task))
				periodStr += fmt.Sprintf(" (Task: %s)", task)
			}

			return q
		})

		// Check if there are any entries
		if len(entries) == 0 {
			pterm.Warning.Println("No time entries found for the selected period")
			return nil
		}

		// Sort entries by start time
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Start.Before(entries[j].Start)
		})

		// Calculate totals
		totalDuration := calculateTotalDuration(entries)
		hoursByDay := calculateHoursByDay(entries)
		hoursByProject := calculateHoursByProject(entries)
		entriesByDay := groupEntriesByDay(entries)
		entriesByProject := groupEntriesByProject(entries)

		// Display report header
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithMargin(10).WithFullWidth().Println(
			fmt.Sprintf("Time Report: %s (%s - %s)",
				periodStr,
				startDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"),
			),
		)

		// Display summary box
		summaryBox := displaySummaryBox(entries, totalDuration, startDate, endDate)

		// Display hours by day with progress bars
		hoursByDayChart := displayHoursByDay(hoursByDay, totalDuration)

		// Display hours by project with bar chart
		hoursByProjectChart := displayHoursByProject(hoursByProject, totalDuration)

		// Display time entries by day
		entriesByDayChart := displayEntriesByDay(entriesByDay)

		// Display time entries by project
		entriesByProjectChart := displayEntriesByProject(entriesByProject)

		panels := pterm.Panels{
			{{Data: summaryBox}, {Data: hoursByProjectChart}},
		}

		if cmd.Bool("today") {
			panels = append(panels, []pterm.Panel{
				{Data: entriesByProjectChart},
			})
		} else {
			panels = append(panels, []pterm.Panel{
				{Data: hoursByDayChart}, {Data: entriesByDayChart},
				{Data: entriesByProjectChart},
			})
		}

		pterm.DefaultPanel.WithPanels(panels).Render()

		return nil
	},
}

// Helper functions

func calculateTotalDuration(entries []*s.TimeEntry) time.Duration {
	var total time.Duration
	for _, entry := range entries {
		total += entry.End.Sub(entry.Start)
	}
	return total
}

func calculateHoursByDay(entries []*s.TimeEntry) map[string]time.Duration {
	hoursByDay := make(map[string]time.Duration)
	for _, entry := range entries {
		day := entry.Start.Format("2006-01-02")
		hoursByDay[day] += entry.End.Sub(entry.Start)
	}
	return hoursByDay
}

func calculateHoursByProject(entries []*s.TimeEntry) map[string]time.Duration {
	hoursByProject := make(map[string]time.Duration)
	for _, entry := range entries {
		hoursByProject[entry.Project] += entry.End.Sub(entry.Start)
	}
	return hoursByProject
}

func groupEntriesByDay(entries []*s.TimeEntry) map[string][]*s.TimeEntry {
	entriesByDay := make(map[string][]*s.TimeEntry)
	for _, entry := range entries {
		day := entry.Start.Format("2006-01-02")
		entriesByDay[day] = append(entriesByDay[day], entry)
	}
	return entriesByDay
}

func groupEntriesByProject(entries []*s.TimeEntry) map[string][]*s.TimeEntry {
	entriesByProject := make(map[string][]*s.TimeEntry)
	for _, entry := range entries {
		entriesByProject[entry.Project] = append(entriesByProject[entry.Project], entry)
	}
	return entriesByProject
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours == 0 && minutes == 0 {
		return "0h 0m"
	} else if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	} else if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func displaySummaryBox(entries []*s.TimeEntry, totalDuration time.Duration, startDate, endDate time.Time) string {
	// Calculate summary data
	uniqueProjects := make(map[string]bool)
	uniqueTasks := make(map[string]bool)

	for _, entry := range entries {
		uniqueProjects[entry.Project] = true
		uniqueTasks[entry.Task] = true
	}

	workingDays := countWorkingDays(startDate, endDate)

	// Create summary panel
	summaryText := fmt.Sprintf(
		"Period: %s to %s\n"+
			"Total Hours: %s\n"+
			"Projects: %d\n"+
			"Tasks: %d\n"+
			"Entries: %d\n"+
			"Working Days: %d\n"+
			"Avg. Working Hours: %.1fh/day",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		formatDuration(totalDuration),
		len(uniqueProjects),
		len(uniqueTasks),
		len(entries),
		workingDays,
		totalDuration.Hours()/float64(workingDays),
	)

	// summary panel
	return pterm.DefaultBox.
		WithTitle("Summary").
		WithTitleTopCenter(true).
		Sprint(summaryText)
}

func countWorkingDays(start, end time.Time) int {
	days := 0
	for date := start; date.Before(end) || date.Equal(end); date = date.AddDate(0, 0, 1) {
		weekday := date.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			days++
		}
	}
	return days
}

func displayHoursByDay(hoursByDay map[string]time.Duration, totalDuration time.Duration) string {
	// Convert map to sorted slice
	type dayHours struct {
		day   string
		hours time.Duration
	}

	days := make([]dayHours, 0, len(hoursByDay))
	for day, hours := range hoursByDay {
		days = append(days, dayHours{day, hours})
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].day < days[j].day
	})

	// Create data for progress bars
	bars := make([]pterm.Bar, len(days))
	for i, dh := range days {
		percentage := 0
		if totalDuration > 0 {
			percentage = int(float64(dh.hours) / float64(totalDuration) * 100)
		}

		// Parse the date to get the weekday name
		t, _ := time.Parse("2006-01-02", dh.day)
		dayName := t.Format("Monday")

		bars[i] = pterm.Bar{
			Label: fmt.Sprintf("%s\n(%s)\n%s", t.Format("01.02"), dayName, formatDuration(dh.hours)),
			Value: percentage,
		}
	}

	// Render the chart
	chart, err := pterm.DefaultBarChart.WithBars(bars).WithShowValue().Srender()
	if err != nil {
		pterm.Error.Println(err)
	}

	box := pterm.DefaultBox.WithTitle("Hours by Day").WithTitleTopCenter(true).Sprint(chart)
	return box
}

func displayHoursByProject(hoursByProject map[string]time.Duration, totalDuration time.Duration) string {
	// Convert map to sorted slice
	type projectHours struct {
		project string
		hours   time.Duration
	}

	projects := make([]projectHours, 0, len(hoursByProject))
	for project, hours := range hoursByProject {
		projects = append(projects, projectHours{project, hours})
	}

	// Sort by hours (descending)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].hours > projects[j].hours
	})

	// Create data for bar chart
	data := pterm.TableData{
		{"Project", "Hours", "Percentage", "Daily Avg."},
	}

	workDays := 5 // Assume 5 working days per week by default

	for _, ph := range projects {
		percentage := 0.0
		if totalDuration > 0 {
			percentage = float64(ph.hours) / float64(totalDuration) * 100
		}

		dailyAvg := ph.hours.Hours() / float64(workDays)

		data = append(data, []string{
			ph.project,
			formatDuration(ph.hours),
			fmt.Sprintf("%.1f%%", percentage),
			fmt.Sprintf("%.1fh", dailyAvg),
		})
	}

	// Render table
	table, err := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	if err != nil {
		pterm.Error.Println(err)
	}

	box := pterm.DefaultBox.WithTitle("Hours by Project").WithTitleTopCenter(true).Sprint(table)

	return box
}

func displayEntriesByDay(entriesByDay map[string][]*s.TimeEntry) string {
	// Get sorted days
	days := make([]string, 0, len(entriesByDay))
	for day := range entriesByDay {
		days = append(days, day)
	}
	sort.Strings(days)

	var content = strings.Builder{}

	// Display entries for each day
	for _, day := range days {
		entries := entriesByDay[day]

		// Convert day to weekday name
		t, _ := time.Parse("2006-01-02", day)
		dayName := t.Format("Monday")

		// Calculate total for the day
		var dayTotal time.Duration
		for _, entry := range entries {
			dayTotal += entry.End.Sub(entry.Start)
		}

		// Create panel title with total hours
		title := fmt.Sprintf("%s (%s) - Total: %s", day, dayName, formatDuration(dayTotal))

		data := pterm.TableData{
			{"Project", "Task", "Duration", "Start", "End"},
		}

		// Create panel content with entries
		for _, entry := range entries {
			duration := entry.End.Sub(entry.Start)

			data = append(data, []string{
				entry.Project,
				entry.Task,
				formatDuration(duration),
				entry.Start.Format("15:04"),
				entry.End.Format("15:04"),
			})
		}

		box := pterm.DefaultBox.WithTitle(title).
			WithTitleTopCenter(true).
			Sprint(pterm.DefaultTable.WithHasHeader().
				WithData(data).
				Srender(),
			)

		content.WriteString(box)
		content.WriteString("\n")
	}

	return content.String()
}

func displayEntriesByProject(entriesByProject map[string][]*s.TimeEntry) string {
	// Get sorted projects
	projects := make([]string, 0, len(entriesByProject))
	for project := range entriesByProject {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	var content = strings.Builder{}

	// Display entries for each project
	for _, project := range projects {
		entries := entriesByProject[project]

		// Calculate total for the project
		var projectTotal time.Duration
		for _, entry := range entries {
			projectTotal += entry.End.Sub(entry.Start)
		}

		// Group entries by task
		entriesByTask := make(map[string][]*s.TimeEntry)
		for _, entry := range entries {
			entriesByTask[entry.Task] = append(entriesByTask[entry.Task], entry)
		}

		// Create table for each task
		data := pterm.TableData{
			{"Task", "Duration", "Start", "End", "Day"},
		}

		// Get sorted tasks
		tasks := make([]string, 0, len(entriesByTask))
		for task := range entriesByTask {
			tasks = append(tasks, task)
		}
		sort.Strings(tasks)

		for _, task := range tasks {
			taskEntries := entriesByTask[task]
			var taskTotal time.Duration

			// Sort entries by start time
			sort.Slice(taskEntries, func(i, j int) bool {
				return taskEntries[i].Start.Before(taskEntries[j].Start)
			})

			for _, entry := range taskEntries {
				duration := entry.End.Sub(entry.Start)
				taskTotal += duration
				data = append(data, []string{
					task,
					formatDuration(duration),
					entry.Start.Format("01.02. 15:04"),
					entry.End.Format("01.02. 15:04"),
					entry.Start.Format("Mon"),
				})
			}
		}

		box := pterm.DefaultBox.WithTitle(fmt.Sprintf("%s - Total: %s", project, formatDuration(projectTotal))).
			WithTitleTopCenter().
			Sprint(pterm.DefaultTable.WithHasHeader().
				WithData(data).Srender(),
			)

		content.WriteString(box)
		content.WriteString("\n")
	}

	return content.String()
}
