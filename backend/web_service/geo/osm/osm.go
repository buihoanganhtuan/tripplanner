package osm

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/paulmach/osm"
)

type OsmGeo struct {
	Domain       domain.Domain
	GeospatialDb *sql.DB
	GraphDb      *sql.DB
}

type OsmNodeId int64
type OsmWayId int64
type OsmRelationId int64

const (
	latBits = 20
	lonBits = 21
)

// Contract implementation

func (o *OsmGeo) GeoPoint(id domain.GeoPointId) (domain.GeoPoint, error) {
	row := o.OsmDb.QueryRow(`SELECT id, lat, lon, level, name FROM  WHERE id = ?`, id)
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
	rows, err := o.OsmDb.Query(`SELECT id, lat, lon, level, name FROM  WHERE id IN $1`, list)
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

func (o *OsmGeo) EdgesTo(id domain.GeoPointId, transport string) ([]domain.GeoEdge, error) {
	rows, err := o.OsmDb.Query(`SELECT id, from, to, edgesCount, travelTime, leftChild, rightChild, `+transport+` WHERE to = ?`, id)
	if err != nil {
		return nil, err
	}
	var ret []domain.GeoEdge
	for rows.Next() {
		var e domain.GeoEdge
		var ttfStr string
		var ttf []domain.TravelTime
		if err = rows.Scan(&e.Id, &e.From, &e.To, &e.OriginalEdges, &ttfStr, e.LeftChild, e.RightChild); err != nil {
			return nil, err
		}
		if ttf, err = parseTTF(ttfStr); err != nil {
			return nil, err
		}
		e.TravelTimeFunction = ttf
	}
}

// sometimes a bus stop node is not connected to any way in/out. This function
// aims to remedy that by binding it to the nearest platform (way), which is supposedly
// connected to the outside world
func (o *OsmGeo) nearestBusPlatforms(relId OsmRelationId, wayId OsmWayId, lim float64) (int64, error) {
	// get all nodes belonging to the same relation
	rows, err := o.GeospatialDb.Query(`
		SELECT
			g.id AS id, ST_Distance(g.geom::geography, t.geom::geography) AS dist
		FROM
			(
				SELECT 
					n.id AS id, p.wkb_geometry AS geom
				FROM
					(
						SELECT
							member_id AS id
						FROM
							osm.relation_members
						WHERE
							relation_id = $1 AND member_type = 'N' AND (member_role = 'stop' OR member_role = 'platform')
					) n
				JOIN
					geospatial.points p
				ON
					n.id = NULLIF(p.osm_id, '')::bigint
			) g
		CROSS JOIN
			(SELECT wkb_geometry AS geom FROM geospatial.lines WHERE osm_id = $2) t
		WHERE
			ST_DWithin(g.geom::geography, t.geom::geography, $3)
		ORDER BY
			ST_Distance(g.geom::geography, t.geom::geography)
		LIMIT
			1
	`, int64(relId), int64(wayId), lim)

	if err != nil {
		return 0, err
	}
	var resId int64 = -1
	var dist float64
	for rows.Next() {
		rows.Scan(&resId, &dist)
	}
	if rows.Err() != nil {
		return 0, rows.Err()
	}
	return resId, nil
}

func (o *OsmGeo) nearestEntrance(nodeId OsmNodeId, subway bool, lim float64) (int64, error) {
	entKw := "train_station_entrance"
	if subway {
		entKw = "subway_entrance"
	}
	rows, err := o.GeospatialDb.Query(`
		SELECT
			osm_id AS id,
		FROM
			geospatial.points
		CROSS JOIN
			(SELECT geom FROM osm.nodes WHERE id = $1)
		WHERE
			other_tags @> '"railway"=>"$2"'::hstore AND ST_DWithin(wkb_geometry::geography, geom::geography, $3)
		ORDER BY
			ST_Distance(wkb_geometry::geography, geom)
		LIMIT 1
	`, int64(nodeId), entKw, lim)

	if err != nil {
		return 0, err
	}
	var resId int64 = -1
	for rows.Next() {
		rows.Scan(&resId)
	}
	if rows.Err() != nil {
		return 0, rows.Err()
	}
	if resId == -1 {
		return 0, errors.New(fmt.Sprintf("no exits within %v meters found for node %v", lim, nodeId))
	}
	return resId, nil
}

// helper functions
func isWalkableWay(w *osm.Way) bool {
	tags := w.TagMap()
	// https://taginfo.openstreetmap.org/keys/footway#values
	if fw, ok := tags["footway"]; ok && (fw == "sidewalk" || fw == "crossing" || fw == "access_aisle") {
		return true
	}
	hw, ok := tags["highway"]
	if !ok {
		return false
	}
	if tv, ok := tags["trail_visibility"]; ok {
		return tv == "excellent" || tv == "good" || tv == "intermediate"
	}
	if ss, ok := tags["sac_scale"]; ok {
		return ss == "hiking" || ss == "mountain_hiking" || ss == "demanding_mountain_hiking"
	}
	if ft, ok := tags["foot"]; ok {
		return ft == "yes" || ft == "designated" || ft == "permissive"
	}
	if pt, ok := tags["public_transport"]; ok {
		return pt == "platform"
	}
	// https://taginfo.openstreetmap.org/keys/highway#values
	if hw == "footway" || hw == "residential" || hw == "pedestrian" || hw == "track" || hw == "unclassified" || hw == "path" || hw == "platform" || hw == "corridor" || hw == "steps" || hw == "elevator" {
		return true
	}
	return false
}

func isBusRoute(r *osm.Relation) bool {
	tags := r.TagMap()
	if tp, ok := tags["type"]; !ok || tp != "route" {
		return false
	}
	if rt, ok := tags["route"]; !ok || rt != "bus" && rt != "trolleybus" {
		return false
	}
	// ver1 has problems with bus stop ordering
	if ver, ok := tags["public_transport:version"]; !ok || ver != "2" {
		return false
	}
	return true
}

func isTrainRoute(r *osm.Relation) bool {
	tags := r.TagMap()
	if tp, ok := tags["type"]; !ok || tp != "route" {
		return false
	}
	if rt, ok := tags["route"]; !ok || rt != "train" || rt != "subway" {
		return false
	}
	if ver, ok := tags["public_transport:version"]; !ok || ver != "2" {
		return false
	}
	return true
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	r := 6378.137e3
	a := math.Sin((lat2 - lat1) / 2)
	b := math.Sin((lon2 - lon1) / 2)
	return 2 * r * math.Asin(math.Sqrt(a*a+math.Cos(lat1)*math.Cos(lat2)*b*b))
}

func parseTTF(ttf string) ([]domain.TravelTime, error) {
	ss := strings.Split(ttf, ";")
	var ret []domain.TravelTime
	for _, s := range ss {
		tmp := strings.Split(s, ":")
		time, err := strconv.ParseFloat(tmp[0], 64)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseFloat(tmp[1], 64)
		if err != nil {
			return nil, err
		}
		ret = append(ret, domain.TravelTime{
			At:    time,
			Value: value,
		})
	}
	return ret, nil
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

func stringify[T any](arr []T, convFun func(T) string) []string {
	res := make([]string, len(arr))
	for i := 0; i < len(arr); i++ {
		res[i] = convFun(arr[i])
	}
	return res
}

func crosstrackDist(lat1, lon1, lat2, lon2, lat3, lon3 float64) float64 {
	r := 6371.
	y1 := math.Sin(lon3-lon1) * math.Cos(lat3)
	x1 := math.Cos(lat1)*math.Sin(lat3) - math.Sin(lat1)*math.Cos(lat3)*math.Cos(lat3-lat1)
	b1 := math.Atan2(y1, x1)
	y2 := math.Sin(lon2-lon1) * math.Cos(lat2)
	x2 := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(lat2-lat1)
	b2 := math.Atan2(y2, x2)
	ac := math.Acos(math.Sin(lat1)*math.Sin(lat3) + math.Cos(lat1)*math.Cos(lat3)*math.Cos(lon3-lon1))
	d := math.Asin(math.Sin(ac)*math.Sin(b1-b2)) * r
	return math.Abs(d)
}

func lastLeq(start int64, end int64, target float64, convFun func(int64) float64) int64 {
	var l, r int64 = start, end
	for l < r {
		m := (l + r + 1) / 2
		if convFun(m) <= target {
			l = m
		} else {
			r = m - 1
		}
	}
	return l
}

func firstLeq(start int64, end int64, target float64, convFun func(int64) float64) int64 {
	var l, r int64 = start, end
	for l < r {
		m := (l + r) / 2
		if convFun(m) <= target {
			r = m
		} else {
			l = m + 1
		}
	}
	return l
}
