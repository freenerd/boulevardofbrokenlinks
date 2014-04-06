package main

import (
	_ "github.com/lib/pq"
	"database/sql"

  "log"
)

type db struct {
  url string
  conn *sql.DB
}

func (db *db) connect() error {
  conn, err := sql.Open("postgres", db.url)
  if err != nil {
    return err
  }

  db.conn = conn

  return nil
}

// assure that the database is in correct state
func (db *db) Check() {
  // check table
  var count int
  err := db.conn.QueryRow("select count(*) from pg_class where relname='users_github'").Scan(&count)
  switch {
  case err == sql.ErrNoRows:
      log.Fatal("Postgres database does not exist.")
  case err != nil:
      log.Fatal(err)
  default:
      if count != 1 {
        log.Fatalf("Expected table users_github to be present once, was present %d", count)
      }
  }
}