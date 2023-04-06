package constants

import (
	"database/sql"
	"fmt"

	utils "github.com/buihoanganhtuan/tripplanner/backend/auth_service/_utils"
)

const (
	PQ_HOST_VAR    = "PQ_HOST"
	PQ_PORT_VAR    = "PQ_PORT"
	PQ_USER_VAR    = "PQ_USER"
	PQ_PASS_VAR    = "PQ_PASS"
	PQ_DB_VAR      = "PQ_DB"
	DatetimeFormat = "2020-12-24 23:59:00"
)

var Db *sql.DB

func init() {
	env := utils.EnvironmentVariableMap{}
	env.Fetch(PQ_HOST_VAR, PQ_PORT_VAR, PQ_USER_VAR, PQ_PASS_VAR)
	if env.Err() != nil {
		panic(fmt.Errorf("environment variable error: %v", env.Err()))
	}
	var err error
	Db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled",
		env.Var(PQ_HOST_VAR),
		env.Var(PQ_PORT_VAR),
		env.Var(PQ_USER_VAR),
		env.Var(PQ_PASS_VAR),
		env.Var(PQ_DB_VAR)))

	err = Db.Ping()

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
}
