package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB holds database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDBConnections = 10
const maxIdleDBConnections = 5
const maxDBLifetime = 5 * time.Minute

// ConnectSQL creates a connection pool to the database specified by the dsn (data
// source name). If the connection cannot be opened, or if the database cannot be
// pinged, an error is returned. The connection pool is set to the global var
// `dbConn` and is returned as well.
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)

	// If can't get the database, panic. Can't continue
	if err != nil {
		panic(err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenDBConnections)
	db.SetMaxIdleConns(maxIdleDBConnections)
	db.SetConnMaxLifetime(maxDBLifetime)

	// Set the global connection pool
	dbConn.SQL = db

	// Test the connection
	err = testDB(db)

	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// testDB tries to ping the database and returns an error if there is one.
func testDB(d *sql.DB) error {
	err := d.Ping()

	if err != nil {
		return err
	}

	return nil
}

// NewDatabase creates a new database connection pool to the database
// specified by the dsn (data source name). If the connection cannot be
// opened, or if the database cannot be pinged, an error is returned.
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	// If error in opening connection, return error
	if err != nil {
		return nil, err
	}

	// If DB can't be pinged, return error
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
