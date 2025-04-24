package db

import (
	"log"
	"os"

	"github.com/ostafen/clover/v2"
)

func NewDB() *clover.DB {
	db, err := clover.Open(GetDefaultPath())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func GetDefaultPath() string {
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

	return dbPath
}
