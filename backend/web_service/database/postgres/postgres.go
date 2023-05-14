package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/app"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/encoding/json"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/environment/variables"
)

const (
	datetimeFormat  = "2006-01-02 15:04:05 -0700"
	host            = "PQ_HOST"
	port            = "PQ_PORT"
	username        = "PQ_USERNAME"
	password        = "PQ_PASSWORD"
	webDbName       = "PQ_WEB_DBNAME"
	userTable       = "PQ_USER_TABLE"
	tripTable       = "PQ_TRIP_TABLE"
	pointTable      = "PQ_POINT_TABLE"
	pointAssocTable = "PQ_POINT_ASSOC_TABLE"
	anonTripTable   = "PQ_ANON_TRIP_TABLE"
	edgeTable       = "PQ_EDGE_TABLE"
	geopointTable   = "PQ_GEOPOINT_TABLE"
	wayTable        = "PQ_WAY_TABLE"
)

var webDb *sql.DB
var ev variables.EnvironmentVariableMap

func InitConnection() error {
	ev.Fetch(host, port, username, password, webDbName)
	if ev.Err() != nil {
		return ev.Err()
	}

	var err error
	webDb, err = sql.Open(`postgres`, fmt.Sprintf(`host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled`,
		ev.Var(host),
		ev.Var(port),
		ev.Var(username),
		ev.Var(password),
		ev.Var(webDbName)))
	if err != nil {
		return err
	}
	return nil
}

func GetUser(id string) (app.User, error) {
	var uid, name, jdStr string

	err := webDb.QueryRow(fmt.Sprintf(`select id, name, join_date from %s where id = ?`, ev.Var(userTable)), id).
		Scan(&uid, &name, &jdStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app.User{}, err
		}
		return app.User{}, err
	}

	jd, err := time.Parse(datetimeFormat, jdStr)

	if err != nil {
		return app.User{}, err
	}

	return app.User{
		Id:       domain.UserId(id),
		Name:     name,
		JoinDate: *json.NewJsonDateTime(jd, datetimeFormat),
	}, nil
}
