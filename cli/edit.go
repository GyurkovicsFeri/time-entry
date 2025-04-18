package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ostafen/clover/v2/query"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	store "github.com/gyurkovicsferi/time-tracker/lib/store"
)

var EditCmd = &cli.Command{
	Name:  "edit",
	Usage: "Edit a time entry",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		db := store.NewStore()
		defer db.Close()

		entries := db.GetTimeEntriesQuery(func(q *query.Query) *query.Query {
			return q.Limit(100).Sort(query.SortOption{
				Field:     "start",
				Direction: -1,
			})
		})

		entriesString := make([]string, len(entries))
		entiresByString := make(map[string]*store.TimeEntry)
		for i, entry := range entries {
			entriesString[i] = fmt.Sprintf("%s | %s | %s | %s | %s", entry.ID, entry.Project, entry.Task, entry.Start.Format(time.RFC850), entry.End.Format(time.RFC850))
			entiresByString[entriesString[i]] = entry
		}

		selected, err := pterm.DefaultInteractiveSelect.WithOptions(entriesString).WithDefaultText("Select a time entry").Show()
		if err != nil {
			return err
		}

		selectedEntry := entiresByString[selected]

		tempFile, err := createTempFileToEdit(selectedEntry)
		if err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		editorCmd := exec.Command(editor, tempFile)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		err = editorCmd.Run()
		if err != nil {
			return err
		}

		editedFile, err := os.ReadFile(tempFile)
		if err != nil {
			return err
		}

		editedFileString := string(editedFile)

		editedFileStringLines := strings.Split(editedFileString, "\n")
		newId := findLineWithPrefixAndTrim(editedFileStringLines, "ID:")
		if newId == "" {
			newId = selectedEntry.ID
		}
		if newId != selectedEntry.ID {
			return fmt.Errorf("invalid file content, ID is not the same")
		}

		newProject := findLineWithPrefixAndTrim(editedFileStringLines, "Project:")
		if newProject == "" {
			newProject = selectedEntry.Project
		}

		newTask := findLineWithPrefixAndTrim(editedFileStringLines, "Task:")
		if newTask == "" {
			newTask = selectedEntry.Task
		}

		var newStart time.Time
		newStartRaw := findLineWithPrefixAndTrim(editedFileStringLines, "Start:")
		if newStartRaw == "" {
			newStart = selectedEntry.Start
		} else {
			newStart, err = time.Parse(time.RFC822, newStartRaw)
			if err != nil {
				return err
			}
		}

		var newEnd time.Time
		newEndRaw := findLineWithPrefixAndTrim(editedFileStringLines, "End:")
		if newEndRaw == "" {
			newEnd = selectedEntry.End
		} else {
			newEnd, err = time.Parse(time.RFC822, newEndRaw)
			if err != nil {
				return err
			}
		}

		if newEnd.Before(newStart) {
			return fmt.Errorf("end time is before start time")
		}

		if newProject != selectedEntry.Project {
			pterm.Println(pterm.LightGreen("Project changed from " + pterm.LightRed(selectedEntry.Project) + " to " + pterm.LightRed(newProject)))
		}

		if newTask != selectedEntry.Task {
			pterm.Println(pterm.LightGreen("Task changed from " + pterm.LightRed(selectedEntry.Task) + " to " + pterm.LightRed(newTask)))
		}

		if newStart != selectedEntry.Start {
			oldStart := selectedEntry.Start.Format(time.RFC822)
			pterm.Println(pterm.LightGreen("Start changed from " + pterm.LightRed(oldStart) + " to " + pterm.LightRed(newStartRaw)))
		}

		if newEnd != selectedEntry.End {
			oldEnd := selectedEntry.End.Format(time.RFC822)
			pterm.Println(pterm.LightGreen("End changed from " + pterm.LightRed(oldEnd) + " to " + pterm.LightRed(newEndRaw)))
		}

		editedEntry := store.TimeEntry{
			ID:      selectedEntry.ID,
			Project: newProject,
			Task:    newTask,
			Start:   newStart,
			End:     newEnd,
		}

		db.UpdateTimeEntry(&editedEntry)

		return nil
	},
}

func createTempFileToEdit(entry *store.TimeEntry) (string, error) {
	tempFile, err := os.CreateTemp("", "time-tracker-entry-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	b := strings.Builder{}

	b.WriteString(fmt.Sprintln("# Modify this file to change the time entry"))
	b.WriteString(fmt.Sprintln("# Don't change the ID line"))
	b.WriteString(fmt.Sprintln())
	b.WriteString(fmt.Sprintln("ID:", entry.ID, "# Don't change this line"))
	b.WriteString(fmt.Sprintln("Project:", entry.Project))
	b.WriteString(fmt.Sprintln("Task:", entry.Task))
	b.WriteString(fmt.Sprintln("Start:", entry.Start.Format(time.RFC822)))
	b.WriteString(fmt.Sprintln("End:", entry.End.Format(time.RFC822)))

	str := b.String()

	tempFile.WriteString(str)

	return tempFile.Name(), nil
}

func findLineWithPrefixAndTrim(lines []string, prefix string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return removeLineComments(strings.TrimSpace(line[len(prefix):]))
		}
	}

	return ""
}

func removeLineComments(line string) string {
	return strings.TrimSpace(strings.Split(line, "#")[0])
}
