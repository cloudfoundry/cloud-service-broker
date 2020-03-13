package main

import (
    _ "github.com/denisenkom/go-mssqldb"
    "database/sql"
    "context"
    "log"
	"fmt"
	"os"
	"strconv"
)

var db *sql.DB

var server string
var port int
var user string
var password string
var database string
var query string
var err error
func main() {
	if len(os.Args) < 6 {
		log.Fatal("Usage: psqlcmd <hostname> <port> <username> <password> <database> <qyery>");
	}
	server = os.Args[1]
	port, err = strconv.Atoi(os.Args[2])
	user = os.Args[3]
	password = os.Args[4]
	database = os.Args[5]
	query = os.Args[6]

    // Build connection string
    connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
        server, user, password, port, database)

    var err error

    // Create connection pool
    db, err = sql.Open("sqlserver", connString)
    if err != nil {
        log.Fatal("Error creating connection pool: ", err.Error())
    }
    ctx := context.Background()
    err = db.PingContext(ctx)
    if err != nil {
        log.Fatal(err.Error())
    }
	tsql := fmt.Sprintf(query)

	_, err = db.ExecContext(ctx, tsql)
	if err != nil {
		log.Fatal(err)
	}
}
