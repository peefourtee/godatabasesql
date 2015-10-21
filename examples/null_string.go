package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db := sqlx.MustConnect("sqlite3", ":memory:")

	if err := createFooTable(db); err != nil {
		log.Fatal("couldn't create table: ", err)
	}

	id, err := insertFoo(db, "hello world")
	if err != nil {
		log.Fatal("failed to insert value: ", err)
	}
	log.Print("inserted foo record ", id)

	foo, err := querySingleFoo(db, id)
	if err != nil {
		log.Fatal("failed to query single foo: ", err)
	}
	log.Printf("found single foo: %+v", foo)
}

const createFooTableSQL = `
CREATE TABLE foo (
	id INTEGER PRIMARY KEY,
	value TEXT NOT NULL,
	timestamp DATETIME
)`

const insertFooQuery = `
INSERT INTO foo (value, timestamp)
VALUES ($1, CURRENT_TIMESTAMP)`

// insert will insert the given value into foo, returning the row's id
func insertFoo(db *sqlx.DB, value string) (int, error) {
	if result, err := db.Exec(insertFooQuery, value); err != nil { // HL
		return 0, err
	} else if id, err := result.LastInsertId(); err != nil { // HL
		return 0, err
	} else {
		return int(id), nil
	}
}

func createFooTable(db *sqlx.DB) error {
	_, err := db.Exec(createFooTableSQL)
	return err
}

// START QUERYROW OMIT
func querySingleFoo(db *sqlx.DB, id int) (Foo, error) {
	var dbFoo struct {
		ID        int
		Value     sql.NullString
		Timestamp time.Time
	}
	if err := db.Get(&dbFoo, "SELECT id, value, timestamp FROM foo WHERE id = $1", id); err != nil {
		return Foo{}, err
	} else {
		foo := Foo{
			ID:        dbFoo.ID,
			Value:     dbFoo.Value.String,
			Timestamp: dbFoo.Timestamp,
		}
		return foo, nil
	}
}

// END QUERYROW OMIT
