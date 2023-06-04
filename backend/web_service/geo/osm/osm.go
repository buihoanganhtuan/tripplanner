package osm

import (
	"database/sql"
	"math"
	"strconv"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/datastructure"
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
var (
	walkable   = datastructure.NewDefaultSet[string]("highway=footway", "")
	unwalkable = datastructure.NewDefaultSet[string]()
)

func (o *OsmGeo) handleOsmNode(node osm.Node) {

}

func extractWalkWay(way osm.Way) string {

}

func extractBusRoute(rel osm.Relation) string {
	mems := []osm.Member(rel.Members)

}

func extractTrainRoute(rel osm.Relation) string {

}

// sometimes a train stop node is not connected to any way in/out. This function
// aims to remedy that by binding it to the nearest platform (way), which is supposedly
// connected to the outside world
func (o *OsmGeo) findNearestPlatform(n *osm.Node) (int, error) {
	dlat := 180. / float64(1<<latBits)
	dlon := 360. / float64(1<<lonBits)
	var i, j int64 = int64(n.Lat / dlat), int64(n.Lon / dlon)
	var hh []string
	for di := int64(-1); di <= 1; di++ {
		for dj := int64(-1); dj <= 1; dj++ {
			hh = append(hh, strconv.FormatInt((i+di)+(j+dj)<<latBits, 10))
		}
	}
	rows, err := o.Db.Query(`
	SELECT 
	FROM (SELECT id FROM  WHERE type = platform AND hash IN (?)) t1
	JOIN (SELECT id, from, to FROM  WHERE type = platform) t2
	ON t1.id = t2.from OR t1.id = t2.to`, strings.Join(hh, ","))
	if err != nil {
		return 0, err
	}

	for rows.Next() {

	}
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
	if rt, ok := tags["route"]; !ok || rt != "train" {
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

func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180
}

func rad2deg(rad float64) float64 {
	return rad * 180. / math.Pi
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

func lastLeq[V ~int | ~float64](start int, end int, target V, leqFun func(int, V) bool) int {
	var l, r int = start, end
	for l < r {
		m := (l + r + 1) / 2
		if leqFun(m, target) {
			l = m
		} else {
			r = m - 1
		}
	}
	return l
}
