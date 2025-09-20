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
	Mutex *sync.Mutex
	DB    *sql.DB
}

func newSQLite() sqlite {
	os.MkdirAll(os.Getenv("DBDir"), 0775)

	db, err := sql.Open("sqlite", filepath.Join(os.Getenv("DBDir"), "database.db"))
	if err != nil {
		log.Println("Initialization function of Sqlite.")
		log.Fatal(err)
	}

	return sqlite{
		Mutex: &sync.Mutex{},
		DB:    db,
	}
}

func (sq *sqlite) Execute(callback func(db *sql.DB) error) error {
	sq.Mutex.Lock()
	defer sq.Mutex.Unlock()

	return callback(sq.DB)
}

var (
	Sqlite = newSQLite()
)
