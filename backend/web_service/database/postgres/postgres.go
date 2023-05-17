package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
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

type Postgres struct {
	webDb *sql.DB
	ev    variables.EnvironmentVariableMap
}

func (p *Postgres) InitConnection() error {
	p.ev.Fetch(host, port, username, password, webDbName)
	if p.ev.Err() != nil {
		return p.ev.Err()
	}

	var err error
	p.webDb, err = sql.Open(`postgres`, fmt.Sprintf(`host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled`,
		p.ev.Var(host),
		p.ev.Var(port),
		p.ev.Var(username),
		p.ev.Var(password),
		p.ev.Var(webDbName)))
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgres) GetUser(id string) (domain.User, error) {
	var uid, name, jdStr string

	err := p.webDb.QueryRow(fmt.Sprintf(`select id, name, join_date from %s where id = ?`, p.ev.Var(userTable)), id).
		Scan(&uid, &name, &jdStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, err
		}
		return domain.User{}, err
	}

	jd, err := time.Parse(datetimeFormat, jdStr)

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       domain.UserId(id),
		Name:     name,
		JoinDate: domain.DateTime(jd),
	}, nil
}
