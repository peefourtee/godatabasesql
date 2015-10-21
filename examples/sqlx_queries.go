package main

import (
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

	foos, err := queryAllFoo(db)
	if err != nil {
		log.Fatal("failed to query all foos: ", err)
	}
	log.Printf("foos: %+v", foos)

	foo, err := querySingleFoo(db, id)
	if err != nil {
		log.Fatal("failed to query single foo: ", err)
	}
	log.Printf("found single foo: %+v", foo)
}

const createFooTableSQL = `
CREATE TABLE foo (
	id INTEGER PRIMARY KEY,
	value TEXT,
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

// START SELECT OMIT
type Foo struct {
	ID            int
	Value         string
	SomeTimestamp time.Time `db:"timestamp"` // HL
}

func queryAllFoo(db *sqlx.DB) ([]Foo, error) {
	var foos []Foo
	err := db.Select(&foos, "SELECT id, value, timestamp FROM foo") // HL
	return foos, err
}

// END SELECT OMIT

// START QUERYROW OMIT
func querySingleFoo(db *sqlx.DB, id int) (Foo, error) {
	foo := Foo{}
	err := db.Get(&foo, "SELECT id, value, timestamp FROM foo WHERE id = $1", id) // HL
	return foo, err
}

// END QUERYROW OMIT
