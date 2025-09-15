package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type sqlite struct {
	mutex *sync.Mutex
	db    *sql.DB
}

func newSQLite() sqlite {
	os.MkdirAll(os.Getenv("DBDir"), 0775)

	db, err := sql.Open("sqlite", filepath.Join(os.Getenv("DBDir"), "database.db"))
	if err != nil {
		log.Println("Initialization function of Sqlite.")
		log.Fatal(err)
	}

	return sqlite{
		mutex: &sync.Mutex{},
		db:    db,
	}
}

func (sq *sqlite) Execute(callback func(db *sql.DB) error) error {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	return callback(sq.db)
}

var (
	Sqlite = newSQLite()
)
