package main

import (
	"database/sql"
	"strconv"

	_ "github.com/lib/pq"
)

func createDb(host string, port int, userName, password, database string, maxConn int) *sql.DB {
	db, err := sql.Open("postgres", "host="+host+" port="+strconv.Itoa(port)+" user="+userName+" password="+password+" dbname="+database+" sslmode=disable")
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(maxConn)
	return db
}
