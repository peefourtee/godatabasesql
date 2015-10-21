package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal("failed to create open db: ", err)
	} else if err = db.Ping(); err != nil {
		log.Fatal("failed to communicate with db: ", err)
	}

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

// START EXEC OMIT
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
func insertFoo(db *sql.DB, value string) (int, error) {
	if result, err := db.Exec(insertFooQuery, value); err != nil { // HL
		return 0, err
	} else if id, err := result.LastInsertId(); err != nil { // HL
		return 0, err
	} else {
		return int(id), nil
	}
}

// END EXEC OMIT

func createFooTable(db *sql.DB) error {
	_, err := db.Exec(createFooTableSQL)
	return err
}

type Foo struct {
	ID        int
	Value     string
	Timestamp time.Time
}

// START QUERY OMIT
func queryAllFoo(db *sql.DB) ([]Foo, error) {
	var foos []Foo
	rows, err := db.Query("SELECT id, value, timestamp FROM foo") // HL
	if err != nil {
		return foos, err
	}

	// ensure the connection is released back to to db's pool // HL
	defer rows.Close() // HL

	for rows.Next() { // HL
		// START SCAN OMIT
		foo := Foo{}
		// scan the row based on SELECT's column order // HL
		if err := rows.Scan(&foo.ID, &foo.Value, &foo.Timestamp); err != nil { // HL
			return foos, err
		}
		foos = append(foos, foo)
		// END SCAN OMIT
	}

	// return any error encountered by iterating rows // HL
	return foos, rows.Err() // HL
}

// END QUERY OMIT

// START QUERYROW OMIT
func querySingleFoo(db *sql.DB, id int) (Foo, error) {
	foo := Foo{}
	row := db.QueryRow("SELECT id, value, timestamp FROM foo WHERE id = $1", id) // HL
	err := row.Scan(&foo.ID, &foo.Value, &foo.Timestamp)                         // HL
	return foo, err
}

// END QUERYROW OMIT
