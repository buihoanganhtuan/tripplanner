package constants

import (
	"database/sql"
	"fmt"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
)

const DatetimeFormat = "2020-12-24 23:59:00"

var Db *sql.DB

const (
	PQ_USERNAME_VAR   = "PQ_USERNAME"
	PQ_PASSWORD_VAR   = "PQ_PASSWORD"
	PQ_WEB_DBNAME_VAR = "PQ_WEB_DBNAME"
)

func init() {
	env := utils.EnvironmentVariableMap{}
	env.Fetch(PQ_USERNAME_VAR, PQ_PASSWORD_VAR, PQ_WEB_DBNAME_VAR)

	Db, err := sql.Open("postgres", fmt.Sprintf("username=%s password=%s dbname=%s sslmode=disabled",
		env.Var(PQ_USERNAME_VAR),
		env.Var(PQ_PASSWORD_VAR),
		env.Var(PQ_WEB_DBNAME_VAR)))
	defer Db.Close()
	err = Db.Ping()

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
}
