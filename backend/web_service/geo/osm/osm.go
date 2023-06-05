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
	Domain domain.Domain
	Db     *sql.DB
}

const (
	latBits = 20
	lonBits = 21
)

// Contract implementation

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

func (o *OsmGeo) EdgesTo(id domain.GeoPointId, transport string) ([]domain.GeoEdge, error) {
	rows, err := o.Db.Query(`SELECT id, from, to, edgesCount, travelTime, leftChild, rightChild, `+transport+` WHERE to = ?`, id)
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

// routines
func (o *OsmGeo) handleOsmNode(node osm.Node) {

}

type stop struct {
	stopPosition osm.NodeID
	platform     osm.WayID
	way          osm.WayID
}

func extractWalkWay(way osm.Way) string {

}

func extractRoute(rel osm.Relation) (string, error) {
	mems := []osm.Member(rel.Members)
	var stops []stop
	var i int
	for ; i < len(mems); i++ {
		mem := mems[i]
		if strings.HasPrefix(mem.Role, "stop") {
			if mem.Type != osm.TypeNode {
				return "", errors.New(fmt.Sprintf(`relation %v: member %v with role "stop" is not a node`, rel.ID, mem.FeatureID()))
			}
			stops = append(stops, stop{
				stopPosition: osm.NodeID(mem.Ref),
				platform:     -1,
				way:          -1,
			})
		}
		if strings.HasPrefix(mem.Role, "platform") {
			if mem.Type != osm.TypeWay {
				return "", errors.New(fmt.Sprintf(`relation %v: member %v with role "platform" is not a way`, rel.ID, mem.FeatureID()))
			}
			if len(stops) == 0 {
				return "", errors.New(fmt.Sprintf("relation %v: malformed route format: platform %v is preceded by no stop", rel.ID, mem.FeatureID()))
			}
			stops[len(stops)-1].platform = osm.WayID(mem.Ref)
		}
		if mem.Role == "" {
			break
		}
	}

	for j := 0; i < len(mems); i++ {
		mem := mems[i]
		if mem.Type != osm.TypeWay {
			return "", errors.New(fmt.Sprintf("relation %v: malformed route format: member %v must be a way on the route, but is not of type way", rel.ID, mem.FeatureID()))
		}
		if mem.Role != "" {
			return "", errors.New(fmt.Sprintf("relation %v: malformed route format: member way %v must have an empty role, but has %s instead", rel.ID, osm.WayID(mem.Ref), mem.Role))
		}
		if j >= len(stops) {
			return "", errors.New(fmt.Sprintf("relation %v: malformed route format: member way %v does not correspond to any stop position", rel.ID, osm.WayID(mem.Ref)))
		}
		stops[j].way = osm.WayID(mem.Ref)
		j++
	}

	for _, stop := range stops {

	}
}

// sometimes a bus stop node is not connected to any way in/out. This function
// aims to remedy that by binding it to the nearest platform (way), which is supposedly
// connected to the outside world
func (o *OsmGeo) nearestBusPlatform(n *osm.Node) (int64, error) {
	dlat := 180. / float64(1<<latBits)
	dlon := 360. / float64(1<<lonBits)
	var i, j int64 = int64(n.Lat / dlat), int64(n.Lon / dlon)
	var hh []string
	for di := int64(-2); di <= 2; di++ {
		for dj := int64(-2); dj <= 2; dj++ {
			hh = append(hh, strconv.FormatInt((i+di)+(j+dj)<<latBits, 10))
		}
	}
	rows, err := o.Db.Query(`
	SELECT
		t3.id AS id, t4.id AS from_id, t4.lat AS from_lat, t4.lon AS from_lon, t5.id AS to_id, t5.lat AS to_lat, t5.lon AS to_lon
	FROM
	(
		SELECT DISTINCT
			t2.id AS id, from, to
		FROM 
			(SELECT id FROM graph.vertices WHERE hash IN (?)) t1
		INNER JOIN 
			(SELECT id, from, to FROM graph.edges WHERE type = bus_platform) t2 ON t2.from = t1.id OR t2.to = t1.id
	) t3
	INNER JOIN
		vertices t4 ON t4.id = t3.from
	INNER JOIN
		vertices t5 ON t5.id = t3.to
	`, strings.Join(hh, ","))
	if err != nil {
		return 0, err
	}

	min := math.MaxFloat64
	var resId int64 = -1
	for rows.Next() {
		var fromLat, fromLon, toLat, toLon float64
		var id, fromId, toId int64
		if err = rows.Scan(&id, &fromId, &fromLat, &fromLon, &toId, &toLat, &toLon); err != nil {
			return 0, err
		}
		d := crosstrackDist(fromLat, fromLon, toLat, toLon, n.Lat, n.Lon)
		if min <= d {
			continue
		}
		min = d
		resId = fromId
		if haversine(n.Lat, n.Lon, fromLat, fromLon) > haversine(n.Lat, n.Lon, toLat, toLon) {
			resId = toId
		}
	}
	if rows.Err() != nil {
		return 0, rows.Err()
	}
	if resId == -1 {
		return 0, errors.New(fmt.Sprintf("no nearest platform found for node %v", n.ID))
	}
	if min > 50 {
		return 0, errors.New(fmt.Sprintf("nearest platform for node %s is too far away (node %s at %v meters)", n.ID, resId, int64(min)))
	}
	return resId, nil
}

func (o *OsmGeo) nearestExitGate(n *osm.Node, isSubway bool) (int64, error) {
	dlat := 180. / float64(1<<latBits)
	dlon := 360. / float64(1<<lonBits)
	var i, j int64 = int64(n.Lat / dlat), int64(n.Lon / dlon)

	var hh []string
	for m := i - 5; m <= i+5; m++ {
		for n := j - 5; n <= j+5; n++ {
			hh = append(hh, strconv.FormatInt(i+j<<latBits, 10))
		}
	}
	t := "train_station_entrance"
	if isSubway {
		t = "subway_entrance"
	}
	rows, err := o.Db.Query(`
	SELECT
		id, lat, lon
	FROM
		graph.vertices
	WHERE
		type = ? AND hash IN (?)
	`, t, strings.Join(hh, ","))
	if err != nil {
		return 0, err
	}

	min := math.MaxFloat64
	var resId int64 = -1
	for rows.Next() {
		var lat, lon float64
		var id int64
		if err = rows.Scan(&id, &lat, &lon); err != nil {
			return 0, err
		}
		d := haversine(lat, lon, n.Lat, n.Lon)
		if min <= d {
			continue
		}
		min = d
		resId = id
	}
	if rows.Err() != nil {
		return 0, rows.Err()
	}
	if resId == -1 {
		return 0, errors.New(fmt.Sprintf("no nearest exit found for node %v", n.ID))
	}
	if min > 100 {
		return 0, errors.New(fmt.Sprintf("nearest exit gate for node %s is too far away (node %s at %v meters)", n.ID, resId, int64(min)))
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
