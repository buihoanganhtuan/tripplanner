package osm

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/geo"
)

type Osm struct {
	Dom *domain.Domain
	Db  *sql.DB

	GeohashLen int
}

type hashId int64

func (o *Osm) findTransitNodes(gp domain.GeoPoint, dist float64) ([]domain.GeoPoint, error) {
	lat := gp.Lat
	lon := gp.Lon

	latBits := o.GeohashLen / 2
	lonBits := o.GeohashLen/2 + o.GeohashLen%2
	numLats := int64(1) << int64(latBits)
	numLons := int64(1) << int64(lonBits)
	dlat := float64(180) / float64(numLats)
	dlon := float64(360) / float64(numLons)

	var lo, hi int
	hi = int(lat / dlat)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		mlat := float64(-90) + float64(mid)*dlat
		if geo.Haversine(lat, lon, mlat, lon) >= dist {
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
		if geo.Haversine(lat, lon, mlat, lon) >= dist {
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
		if geo.Haversine(lat, lon, lat, mlon) >= dist {
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
		if geo.Haversine(lat, lon, lat, mlon) >= dist {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	rlon := int64(lo)

	var geoHashes []hashId
	for j := blat; j <= tlat; j++ {
		for k := llon; k <= rlon; k++ {
			geoHashes = append(geoHashes, hashId(j+k<<latBits))
		}
	}

	gpp, err := o.geoPointsWithHashes(geoHashes)
	if err != nil {
		return nil, err
	}

	var tmp []domain.GeoPoint
	for _, gp := range gpp {
		if geo.Haversine(lat, lon, gp.Lat, gp.Lon) > dist {
			continue
		}
		tmp = append(tmp, gp)
	}

	return tmp, nil
}

func (o *Osm) geoPointsWithHashes(hh []hashId) ([]domain.GeoPoint, error) {
	var hs []string
	for _, h := range hh {
		hs = append(hs, strconv.FormatInt(int64(h), 10))
	}
	q := `SELECT Id, Lat, Lon, Name, Address, Tags FROM  WHERE HashId in (?)`
	rows, err := o.Db.Query(q, strings.Join(hs, ","))
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
