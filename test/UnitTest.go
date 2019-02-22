package test

import (
	"database/sql"
	"log"

	"github.com/globalsign/mgo"
)

// MongoDBTestURL ...
const MongoDBTestURL = "127.0.0.1:27017/test"

// PostgreSQLTestURL ...
const PostgreSQLTestURL = "postgres://postgres@127.0.0.1:5432/test?sslmode=disable"

// OpenMongoDBForTest ...
func OpenMongoDBForTest() *mgo.Database {
	session, err := mgo.Dial(MongoDBTestURL)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session.DB("")
}

var postgresDB *sql.DB

// OpenPostgreSQForTest ...
func OpenPostgreSQForTest() *sql.DB {
	if postgresDB != nil {
		return postgresDB
	}
	db, err := sql.Open("postgres", PostgreSQLTestURL)
	if err != nil {
		log.Fatal(err)
	}
	postgresDB = db
	return db
}
