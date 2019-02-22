package storage

import (
	"database/sql"

	"github.com/freeznet/tomato/config"
	"github.com/freeznet/tomato/test"
	_ "github.com/lib/pq" // postgres driver
	"github.com/globalsign/mgo"
)

// OpenMongoDB 打开 MongoDB
func OpenMongoDB() *mgo.Database {
	// 此处仅用于测试
	if config.TConfig.DatabaseURI == "" {
		config.TConfig.DatabaseURI = test.MongoDBTestURL
	}

	session, err := mgo.Dial(config.TConfig.DatabaseURI)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session.DB("")
}

// OpenPostgreSQL 打开 PostgreSQL
func OpenPostgreSQL() *sql.DB {
	db, err := sql.Open("postgres", config.TConfig.DatabaseURI)
	if err != nil {
		panic(err)
	}
	// we use 80% of pg max connections
	db.SetMaxOpenConns(int(float64(config.TConfig.PgMaxConnections) * 0.7))
	db.SetMaxIdleConns(int(float64(config.TConfig.PgMaxConnections) * 0.1))
	return db
}
