package constants

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/environment/variables"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

var Db *sql.DB
var Kvs *redis.Client
var Pk *rsa.PublicKey
var Ev variables.EnvironmentVariableMap

const (
	DatetimeFormat              = "2006-01-02 15:04:05 -0700"
	SqlHostVar                  = "PQ_HOST"
	SqlPortVar                  = "PQ_PORT"
	SqlUsernameVar              = "PQ_USERNAME"
	SqlPasswordVar              = "PQ_PASSWORD"
	SqlWebDbNameVar             = "PQ_WEB_DBNAME"
	SqlUserTableVar             = "PQ_USER_TABLE"
	SqlTripTableVar             = "PQ_TRIP_TABLE"
	SqlPointTableVar            = "PQ_POINT_TABLE"
	SqlPointAssociationTableVar = "PQ_POINT_ASSOC_TABLE"
	SqlAnonTripTableVar         = "PQ_ANON_TRIP_TABLE"
	SqlEdgeTableVar             = "PQ_EDGE_TABLE"
	SqlGeoPointTableVar         = "PQ_GEOPOINT_TABLE"
	SqlWayTableVar              = "PQ_WAY_TABLE"
	KvsHostVar                  = "REDIS_HOST"
	KvsPortVar                  = "REDIS_PORT"
	KvsPasswordVar              = "REDIS_PASSWORD"
	KvsDelHKeyVar               = "REDIS_DELETE_TRANS_KEY"
	PublicKeyPathVar            = "PUBLIC_KEY_PATH"
	AuthServiceName             = "Tripplanner:AuthService"
	WebServiceName              = "Tripplanner:WebService"
	GeohashLen                  = 41
	RouteSearchRadius           = 2000.
	MaxCandidateTrips           = 10
	MaxCandidateAnonTrips       = 4
)

func init() {
	Ev.Fetch(
		SqlHostVar,
		SqlPortVar,
		SqlUsernameVar,
		SqlPasswordVar,
		SqlWebDbNameVar,
		SqlUserTableVar,
		SqlTripTableVar,
		KvsHostVar,
		KvsPortVar,
		KvsPasswordVar,
		KvsDelHKeyVar,
		PublicKeyPathVar)

	var err error

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
	err = Db.Ping()
	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}

	// Load authentication server's public key for access token validation
	Ev.Fetch(PublicKeyPathVar, SqlUserTableVar)
	if Ev.Err() != nil {
		panic(fmt.Errorf("environment variable error: %v", Ev.Err()))
	}

	b, err := os.ReadFile(Ev.Var(PublicKeyPathVar))
	if err != nil {
		panic(fmt.Errorf("cannot read public key file: %v", err))
	}

	Pk, err = jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		panic(fmt.Errorf("fail to parse public key from file: %v", err))
	}

	// Connect to key-value store
	dbn, err := strconv.Atoi(Ev.Var(KvsDelHKeyVar))
	if err != nil {
		panic(fmt.Errorf("cannot parse %v as an int", Ev.Var(KvsDelHKeyVar)))
	}
	KvStore := redis.NewClient(&redis.Options{
		Addr:     Ev.Var(KvsHostVar) + ":" + Ev.Var(KvsPortVar),
		Password: Ev.Var(KvsPasswordVar),
		DB:       dbn,
	})

	ctx := context.Background()
	_, err = KvStore.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("error connecting to key-value store: %v", err))
	}
}
