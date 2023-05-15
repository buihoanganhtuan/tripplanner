package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/app"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
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
	GeohashLen      = 41
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

/*
Get latitude and longitude of bottom left and top right of a square
centered at current position and has side = 2*dist
*/
func GetNearbyPoints(id domain.GeoPointId, dist float64) ([]domain.GeoPoint, error) {
	intId, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return nil, err
	}

	row := webDb.QueryRow(`SELECT Lat, Lon, Name, Address FROM  WHERE Id = ?`, intId)
	var lat, lon float64
	var name, addr string
	if err = row.Scan(&lat, &lon, &name, &addr); err != nil {
		return nil, err
	}

	latBits := GeohashLen / 2
	lonBits := GeohashLen/2 + GeohashLen%2
	numLats := int64(1) << int64(latBits)
	numLons := int64(1) << int64(lonBits)
	dlat := float64(180) / float64(numLats)
	dlon := float64(360) / float64(numLons)

	var lo, hi int
	hi = int(lat / dlat)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		mlat := float64(-90) + float64(mid)*dlat
		if haversine(lat, lon, mlat, lon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	blat := int64(lo)

	lo = int(lat/dlat) + 1
	hi = 1<<latBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		mlat := float64(-90) + float64(mid)*dlat
		if haversine(lat, lon, mlat, lon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	tlat := int64(lo)

	lo = 0
	hi = int(lon / dlon)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		mlon := float64(-180) + float64(mid)*dlon
		if haversine(lat, lon, lat, mlon) >= dist {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	llon := int64(lo)

	lo = int(lon/dlon) + 1
	hi = 1<<lonBits - 1
	for lo < hi {
		mid := (lo + hi) / 2
		mlon := float64(-180) + float64(mid)*dlon
		if haversine(lat, lon, lat, mlon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	rlon := int64(lo)

	var geoHashes []string
	for j := blat; j <= tlat; j++ {
		for k := llon; k <= rlon; k++ {
			geoHashes = append(geoHashes, strconv.FormatInt(j+k<<latBits, 10))
		}
	}

	q := `SELECT GeoPointId, Lat, Lon, Name, Address, Tags FROM  WHERE GeoHashId in (?)`
	rows, err := webDb.Query(q, strings.Join(geoHashes, ","))
	if err != nil {
		return nil, err
	}

	var tmp []domain.GeoPoint
	for rows.Next() {
		var nlat, nlon float64
		var id, name, addr, tags string
		if err = rows.Scan(&id, &nlat, &nlon, &name, &addr, &tags); err != nil {
			return nil, err
		}
		if haversine(lat, lon, nlat, nlon) > dist {
			continue
		}

		var tokens []string
		// handle address
		tokens = strings.Split(addr, " ")
		var add domain.Address
		if len(tokens) < 4 {
			continue
		}
		add.Prefecture = tokens[0]
		add.City = tokens[1]
		add.District = tokens[2]
		if len(tokens) >= 4 {
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
		tmp = append(tmp, domain.GeoPoint{
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

	return tmp, nil
}

func FindShortestPath(src datastructure.Set[RouteId], dst datastructure.Set[RouteId], preferMode string, budget int) []RouteId {

}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}
