package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type sqlite struct {
	Mutex *sync.Mutex
	DB    *sqlx.DB
}

func newSQLite() sqlite {
	dbDir := "./database"
	os.Mkdir(dbDir, 0755)

	db, err := sqlx.Connect("sqlite", filepath.Join(dbDir, "database.db"))
	if err != nil {
		log.Println("Initialization function of Sqlite.")
		log.Fatal(err)
	}

	return sqlite{
		Mutex: &sync.Mutex{},
		DB:    db,
	}
}

func (sq *sqlite) Execute(callback func(db *sqlx.DB) error) error {
	sq.Mutex.Lock()
	defer sq.Mutex.Unlock()

	return callback(sq.DB)
}

var (
	Sqlite = newSQLite()
)
