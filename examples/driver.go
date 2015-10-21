package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:") // HL
	if err != nil {
		panic("failed to create open db: " + err.Error())
	} else if err = db.Ping(); err != nil { // HL
		panic("failed to communicate with db: " + err.Error())
	}
}
