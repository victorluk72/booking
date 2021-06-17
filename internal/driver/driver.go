package driver

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	//Drivers for Postgres
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB holds database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

// Basic setting to DB
const maxOpenDBConn = 100
const maxIdleDBConn = 5
const maxDBLifeTime = 5 * time.Minute

// ConnectSQL creats database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	//Set up some default parameters for connections
	db.SetMaxOpenConns(maxOpenDBConn)
	db.SetMaxIdleConns(maxIdleDBConn)
	db.SetConnMaxLifetime(maxDBLifeTime)

	dbConn.SQL = db

	//test db connection again using my custom func testDB()
	err = testDB(db)
	if err != nil {
		return nil, err
	}

	//no erro scenario
	return dbConn, nil

}

//NewDatabese creates a new databaseinstance for our app
func NewDatabase(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to connect: %v\n", err))
		return nil, err
	}

	//Ping the connection
	if err = conn.Ping(); err != nil {
		log.Fatal("Cannot ping database")
		return nil, err

	}
	fmt.Println("Pinged database...")

	return conn, err
}

// test DB tries to ping db connection
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}

	return nil

}
