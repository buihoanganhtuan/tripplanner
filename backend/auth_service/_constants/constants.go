package constants

import (
	"database/sql"
	"fmt"
)

const DatetimeFormat = "2020-12-24 23:59:00"

var Db *sql.DB

func init() {
	Db, err := sql.Open("postgres", fmt.Sprintf("username=%s password=%s dbname=%s sslmode=disabled",
		env.Var("PQ_USERNAME"),
		env.Var("PQ_PASSWORD"),
		env.Var("PQ_AUTH_DBNAME")))
	defer Db.Close()
	err = Db.Ping()

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
}
