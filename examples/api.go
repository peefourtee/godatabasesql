package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/schema"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Foo struct {
	ID            int
	Value         string
	SomeTimestamp time.Time `db:"timestamp"` // HL
}

type Page struct {
	Size   int `schema:"page_size"`
	Number int `schema:"page"`
}

func (p Page) Statement() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Size, p.Size*p.Number)
}

type FooListOptions struct {
	IDs   []int `schema:"id"`
	Value string
	Page
}

func (o FooListOptions) Wheres(wheres []string, params []interface{}) ([]string, []interface{}) {
	if len(o.IDs) > 0 {
		idIn := make([]string, len(o.IDs))
		for i, id := range o.IDs {
			params = append(params, id)
			idIn[i] = fmt.Sprintf("?")
		}
		idWhere := fmt.Sprintf("foo.id IN (%s)", strings.Join(idIn, ","))
		wheres = append(wheres, idWhere)
	}
	if o.Value != "" {
		params = append(params, o.Value)
		wheres = append(wheres, fmt.Sprintf("foo.value = $%d", len(params)))
	}

	return wheres, params
}

type Queryer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Rebind(string) string
	Select(dest interface{}, query string, args ...interface{}) error
}

type FooStore struct {
	db Queryer
}

func (s FooStore) List(opts FooListOptions) ([]Foo, error) {
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = opts.Wheres(wheres, params)

	query := "SELECT id, value, timestamp FROM foo"
	if len(wheres) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(wheres, " AND "))
	}

	if opts.Page.Size > 0 {
		query += " " + opts.Page.Statement()
	}

	foos := make([]Foo, 0)
	return foos, s.db.Select(&foos, s.db.Rebind(query), params...)
}

func fooListEndpoint(w http.ResponseWriter, r *http.Request, store *FooStore) {
	opts := FooListOptions{}
	if err := schema.NewDecoder().Decode(&opts, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	foos, err := store.List(opts)
	if err != nil {
		log.Print("failed to perform query: ", err)
		http.Error(w, "failed to perform query", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(foos); err != nil {
		log.Print("failed to write response to %+v: %s", opts, err)
	}
}

var addr = flag.String("addr", "127.0.0.1:8080", "address to bind http server to")

func main() {
	db := sqlx.MustConnect("sqlite3", ":memory:")

	if err := createFooTable(db); err != nil {
		log.Fatal("couldn't create table: ", err)
	}

	for i := 0; i < 10; i++ {
		id, err := insertFoo(db, "hello world "+strconv.Itoa(i))
		if err != nil {
			log.Fatal("failed to insert value: ", err)
		}
		log.Print("inserted foo record ", id)
	}

	h := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fooListEndpoint(w, r, &FooStore{db: db})
	}))
	h = handlers.LoggingHandler(os.Stdout, h)
	h = handlers.ContentTypeHandler(h, "application/json")

	http.Handle("/foo/", h)

	flag.Parse()
	log.Print("starting http server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Printf("http server failed: ", err)
	}
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
