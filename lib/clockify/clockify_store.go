package clockify

import (
	"log"

	"github.com/gyurkovicsferi/time-tracker/lib/store"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

const (
	ClockifyTimeEntryCollection = "clockify_time_entries"
	ClockifyConfigCollection    = "clockify_config"
)

type ClockifyStore struct {
	db *clover.DB
}

type ClockifyConfig struct {
	ID          string `clover:"id"`
	WorkspaceID string `clover:"workspace_id"`
	APIKey      string `clover:"api_key"`
}

type ClockifyTimeEntry struct {
	ID          string `clover:"id"`
	TimeEntryID string `clover:"time_entry_id"`
	ClockifyID  string `clover:"clockify_id"`
	Deleted     bool   `clover:"deleted"`
}

func NewClockifyStore(cloverDB *clover.DB) *ClockifyStore {
	store := &ClockifyStore{
		db: cloverDB,
	}
	store.createCollectionIfNotExists()
	store.createConfigCollectionIfNotExists()
	return store
}

func (s *ClockifyStore) createCollectionIfNotExists() {
	hasCollection, err := s.db.HasCollection(ClockifyTimeEntryCollection)
	if err != nil {
		log.Fatal(err)
	}

	if !hasCollection {
		err = s.db.CreateCollection(ClockifyTimeEntryCollection)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *ClockifyStore) createConfigCollectionIfNotExists() {
	hasCollection, err := s.db.HasCollection(ClockifyConfigCollection)
	if err != nil {
		log.Fatal(err)
	}

	if !hasCollection {
		err = s.db.CreateCollection(ClockifyConfigCollection)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func (s *ClockifyStore) InsertTimeEntry(timeEntry *store.TimeEntry, clockifyID string) error {
	clockifyTimeEntry := &ClockifyTimeEntry{
		TimeEntryID: timeEntry.ID,
		ClockifyID:  clockifyID,
		Deleted:     false,
	}

	doc := document.NewDocumentOf(clockifyTimeEntry)
	return s.db.Insert(ClockifyTimeEntryCollection, doc)
}

func (s *ClockifyStore) MakeClockifyTimeEntryDeleted(timeEntryID string) error {
	return s.db.Update(query.NewQuery(ClockifyTimeEntryCollection).
		Where(query.Field("time_entry_id").
			Eq(timeEntryID),
		),
		map[string]interface{}{
			"deleted": true,
		},
	)
}

func (s *ClockifyStore) GetClockifyTimeEntry(timeEntry *store.TimeEntry) (*ClockifyTimeEntry, error) {
	doc, err := s.db.FindFirst(query.NewQuery(ClockifyTimeEntryCollection).
		Where(query.Field("time_entry_id").
			Eq(timeEntry.ID),
		),
	)

	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, nil
	}

	return unmarshalClockifyTimeEntry(doc), nil
}

func unmarshalClockifyTimeEntry(doc *document.Document) *ClockifyTimeEntry {
	clockifyTimeEntry := &ClockifyTimeEntry{}
	err := doc.Unmarshal(clockifyTimeEntry)
	if err != nil {
		log.Fatal(err)
	}
	return clockifyTimeEntry
}

func (s *ClockifyStore) GetClockifyConfig() (*ClockifyConfig, error) {
	doc, err := s.db.FindFirst(query.NewQuery(ClockifyConfigCollection))
	if err != nil {
		return nil, err
	}

	return unmarshalClockifyConfig(doc), nil
}

func unmarshalClockifyConfig(doc *document.Document) *ClockifyConfig {
	clockifyConfig := &ClockifyConfig{}
	err := doc.Unmarshal(clockifyConfig)
	if err != nil {
		log.Fatal(err)
	}
	return clockifyConfig
}

func (s *ClockifyStore) InsertClockifyConfig(config *ClockifyConfig) error {
	// Delete all existing configs
	err := s.DeleteClockifyConfig()
	if err != nil {
		return err
	}

	// Insert the new config
	doc := document.NewDocumentOf(config)
	return s.db.Insert(ClockifyConfigCollection, doc)
}

func (s *ClockifyStore) DeleteClockifyConfig() error {
	return s.db.Delete(query.NewQuery(ClockifyConfigCollection))
}
