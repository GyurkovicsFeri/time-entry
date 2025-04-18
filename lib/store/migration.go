package store

import (
	"log"
)

func (s *Store) Migrate() {
	s.createTimeEntryCollectionIfNotExists()
	s.createCurrentTimeEntryCollectionIfNotExists()
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
