package store

import (
	"log"
	"os"
)

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
