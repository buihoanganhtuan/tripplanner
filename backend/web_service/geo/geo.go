package geo

import (
	"database/sql"
	"math"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
)

type OsmGeo struct {
	Domain domain.Domain
	Db     *sql.DB
}

func (o *OsmGeo) GeoPoint(id domain.GeoPointId) (domain.GeoPoint, error) {
	row := o.Db.QueryRow(`SELECT id, lat, lon, level, name FROM  WHERE id = ?`, id)
	var p domain.GeoPoint
	if err := row.Scan(&p.Id, &p.Lat, &p.Lon, &p.Level, &p.Name); err != nil {
		return p, err
	}
	return p, nil
}

func (o *OsmGeo) GeoPoints(ids []domain.GeoPointId) ([]domain.GeoPoint, error) {
	var ss []string
	for _, id := range ids {
		ss = append(ss, string(id))
	}
	list, err := parenthesize(ss, '(', ')', ',')
	if err != nil {
		return nil, err
	}
	rows, err := o.Db.Query(`SELECT id, lat, lon, level, name FROM  WHERE id IN ?`, list)
	var ret []domain.GeoPoint
	for rows.Next() {
		var p domain.GeoPoint
		if err = rows.Scan(&p.Id, &p.Lat, &p.Lon, &p.Level, &p.Name); err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}

func parenthesize(ss []string, left rune, right rune, sep rune) (string, error) {
	var sb strings.Builder
	if _, err := sb.WriteRune(left); err != nil {
		return "", err
	}
	for i, s := range ss {
		if _, err := sb.WriteString(s); err != nil {
			return "", err
		}
		if i == len(ss)-1 {
			continue
		}
		if _, err := sb.WriteRune(sep); err != nil {
			return "", err
		}
	}
	if _, err := sb.WriteRune(right); err != nil {
		return "", err
	}
	return sb.String(), nil
}
