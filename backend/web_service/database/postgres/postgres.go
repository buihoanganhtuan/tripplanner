package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (p *Postgres) GeoGeoPointsWithHashes(hh []domain.GeoHashId) ([]domain.GeoPoint, error) {
	var hs []string
	for _, h := range hh {
		hs = append(hs, string(h))
	}
	q := `SELECT Id, Lat, Lon, Name, Address, Tags FROM  WHERE HashId in (?)`
	rows, err := p.webDb.Query(q, strings.Join(hs, ","))
	if err != nil {
		return nil, err
	}

	var res []domain.GeoPoint
	for rows.Next() {
		var id, name, addr, tags string
		var lat, lon float64

		var tokens []string
		tokens = strings.Split(addr, " ")
		var add domain.Address
		if len(tokens) > 0 {
			add.Prefecture = tokens[0]
		}
		if len(tokens) > 1 {
			add.City = tokens[1]
		}
		if len(tokens) > 2 {
			add.District = tokens[2]
		}
		if len(tokens) > 3 {
			add.LandNumber = tokens[3]
		}

		// handletags
		var t []domain.KeyValuePair
		for _, kv := range strings.Split(tags, ";") {
			tokens = strings.Split(kv, ":")
			if len(tokens) != 2 {
				continue
			}
			t = append(t, domain.KeyValuePair{
				Key:   tokens[0],
				Value: tokens[1],
			})
		}
		res = append(res, domain.GeoPoint{
			Id:      domain.GeoPointId(id),
			Lat:     lat,
			Lon:     lon,
			Name:    &name,
			Address: add,
			Tags:    t,
		})
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}
