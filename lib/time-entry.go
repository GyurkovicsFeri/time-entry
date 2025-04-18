package timeentry

import (
	"time"

	"github.com/google/uuid"
	s "github.com/gyurkovicsferi/time-tracker/lib/store"
)

func Start(project, task string, store *s.Store) *s.CurrentTimeEntry {
	return NewCurrentTimeEntry(store, project, task, time.Now())
}

func NewCurrentTimeEntry(store *s.Store, project, task string, start time.Time) *s.CurrentTimeEntry {
	if store == nil {
		store = s.NewStore()
		defer store.Close()
	}

	current := store.GetCurrentTimeEntry()

	if current != nil {
		Stop(store, current, time.Now())
	}

	currentTimeEntry := &s.CurrentTimeEntry{
		ID:      uuid.New().String(),
		Project: project,
		Task:    task,
		Start:   start,
	}

	store.InsertCurrentTimeEntry(currentTimeEntry)
	return currentTimeEntry
}

func Stop(store *s.Store, currentTimeEntry *s.CurrentTimeEntry, end time.Time) *s.TimeEntry {
	timeEntry := &s.TimeEntry{
		ID:      currentTimeEntry.ID,
		Project: currentTimeEntry.Project,
		Task:    currentTimeEntry.Task,
		Start:   currentTimeEntry.Start,
		End:     end,
	}

	store.InsertTimeEntry(timeEntry)
	store.DeleteCurrentTimeEntry()

	return timeEntry
}

func GetProjects(store *s.Store) []string {
	return store.GetProjects()
}

func GetTasks(store *s.Store, project string) []string {
	return store.GetTasks(project)
}
