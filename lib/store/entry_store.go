package store

import (
	"log"
	"os"
	"time"

	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

const (
	CurrentTimeEntryCollection = "current-time-entry"
	TimeEntryCollection        = "time-entry"
)

type CurrentTimeEntry struct {
	ID      string    `clover:"id"`
	Project string    `clover:"project"`
	Task    string    `clover:"task"`
	Start   time.Time `clover:"start"`
}

type TimeEntry struct {
	ID      string    `clover:"id"`
	Project string    `clover:"project"`
	Task    string    `clover:"task"`
	Start   time.Time `clover:"start"`
	End     time.Time `clover:"end"`
}

type Store struct {
	db *clover.DB
}

func NewStore() *Store {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := homeDir + "/.time-tracker/db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = os.MkdirAll(dbPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	db, err := clover.Open(homeDir + "/.time-tracker/db")
	if err != nil {
		log.Fatal(err)
	}

	store := &Store{
		db: db,
	}
	store.createTimeEntryCollectionIfNotExists()
	store.createCurrentTimeEntryCollectionIfNotExists()
	return store
}

func (s *Store) createTimeEntryCollectionIfNotExists() {
	hasCollection, err := s.db.HasCollection(TimeEntryCollection)
	if err != nil {
		log.Fatal(err)
	}

	if !hasCollection {
		err = s.db.CreateCollection("time-entry")
		if err != nil {
			log.Fatal(err)
		}
	}

	hasIndex, err := s.db.HasIndex(TimeEntryCollection, "project")
	if err != nil {
		log.Fatal(err)
	}

	if !hasIndex {
		err = s.db.CreateIndex(TimeEntryCollection, "project")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Store) createCurrentTimeEntryCollectionIfNotExists() {
	hasCollection, err := s.db.HasCollection(CurrentTimeEntryCollection)
	if err != nil {
		log.Fatal(err)
	}

	if !hasCollection {
		err = s.db.CreateCollection(CurrentTimeEntryCollection)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Store) InsertCurrentTimeEntry(currentTimeEntry *CurrentTimeEntry) string {
	doc := document.NewDocumentOf(currentTimeEntry)

	id, err := s.db.InsertOne(CurrentTimeEntryCollection, doc)
	if err != nil {
		log.Fatal(err)
	}

	return id
}

func (s *Store) DeleteCurrentTimeEntry() {
	err := s.db.Delete(query.NewQuery(CurrentTimeEntryCollection))
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Store) InsertTimeEntry(timeEntry *TimeEntry) string {
	doc := document.NewDocumentOf(timeEntry)

	id, err := s.db.InsertOne(TimeEntryCollection, doc)
	if err != nil {
		log.Fatal(err)
	}

	return id
}

func (s *Store) GetCurrentTimeEntry() *CurrentTimeEntry {
	doc, err := s.db.FindFirst(query.NewQuery(CurrentTimeEntryCollection))
	if err != nil {
		log.Fatal(err)
	}

	if doc == nil {
		return nil
	}

	currentTimeEntry := &CurrentTimeEntry{}
	err = doc.Unmarshal(currentTimeEntry)
	if err != nil {
		log.Fatal(err)
	}
	return currentTimeEntry
}

func (s *Store) GetTimeEntries() []*TimeEntry {
	docs, err := s.db.FindAll(query.NewQuery(TimeEntryCollection))
	if err != nil {
		log.Fatal(err)
	}

	timeEntries := make([]*TimeEntry, len(docs))
	for i, doc := range docs {

		timeEntries[i] = unmarshalTimeEntry(doc)
	}

	return timeEntries
}

func unmarshalTimeEntry(doc *document.Document) *TimeEntry {
	timeEntry := &TimeEntry{}
	err := doc.Unmarshal(timeEntry)
	if err != nil {
		log.Fatal(err)
	}
	return timeEntry
}

func (s *Store) GetTimeEntryForToday() *TimeEntry {
	docs, err := s.db.FindAll(
		query.NewQuery(TimeEntryCollection).
			Where(query.Field("start").
				GtEq(StartOfDay(time.Now())).
				And(
					query.Field("start").
						LtEq(EndOfDay(time.Now())),
				),
			),
	)
	if err != nil {
		log.Fatal(err)
	}

	timeEntry := &TimeEntry{}
	err = docs[0].Unmarshal(timeEntry)
	if err != nil {
		log.Fatal(err)
	}
	return timeEntry
}

func (s *Store) Close() {
	err := s.db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999, t.Location())
}

func (store *Store) GetProjects() []string {
	docs, err := store.db.FindAll(query.NewQuery(TimeEntryCollection).Limit(100))
	if err != nil {
		log.Fatal(err)
	}

	projects := make([]string, len(docs))
	for i, doc := range docs {
		projects[i] = doc.Get("project").(string)
	}
	return projects
}

func (store *Store) GetTasks(project string) []string {
	docs, err := store.db.FindAll(query.NewQuery(TimeEntryCollection).Where(query.Field("project").Eq(project)).Limit(10))
	if err != nil {
		log.Fatal(err)
	}

	tasks := make([]string, len(docs))
	for i, doc := range docs {
		tasks[i] = doc.Get("task").(string)
	}
	return tasks
}
