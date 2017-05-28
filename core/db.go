package core

import (
	"database/sql"

	// Needed since we are using this for opening the connection
	_ "github.com/go-sql-driver/mysql"
)

// DB class to work with MySQL database
type DB struct {
	DBConnection *sql.DB
	mysqlServer  string
	mysqlUser    string
	mysqlDB      string
	mysqlPw      string
}

// SetMysqlServer allows setting the MySQL server to use
func (db *DB) SetMysqlServer(mysqlServer string) {
	db.mysqlServer = mysqlServer
}

// SetMysqlUser allows setting the MySQL user to use
func (db *DB) SetMysqlUser(mysqlUser string) {
	db.mysqlUser = mysqlUser
}

// SetMysqlDB allows setting the MySQL database to use
func (db *DB) SetMysqlDB(mysqlDB string) {
	db.mysqlDB = mysqlDB
}

// SetMysqlPw allows setting the MySQL password to use
func (db *DB) SetMysqlPw(mysqlPw string) {
	db.mysqlPw = mysqlPw
}

// Connect to the MySQL server
func (db *DB) Connect() error {
	var err error
	db.DBConnection, err = sql.Open("mysql", db.mysqlUser+":"+db.mysqlPw+"@tcp("+db.mysqlServer+")/"+db.mysqlDB)
	if err != nil {
		return err
	}

	err = db.DBConnection.Ping()
	return err
}

// New will create a database connection and return the sql.DB
func (db *DB) New(mysqlServer string, mysqlDB string, mysqlUser string, mysqlPw string) (*sql.DB, error) {
	db.SetMysqlServer(mysqlServer)
	db.SetMysqlDB(mysqlDB)
	db.SetMysqlUser(mysqlUser)
	db.SetMysqlPw(mysqlPw)
	err := db.Connect()
	return db.DBConnection, err
}
