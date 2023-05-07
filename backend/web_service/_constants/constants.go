package constants

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

var Db *sql.DB
var Kvs *redis.Client
var Pk *rsa.PublicKey
var Ev utils.EnvironmentVariableMap

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
)

// Tagging to assist JSON Marshalling (converting a structured data into a JSON string)
// Note that in this case, missing, null, and zero values are handled in the same way so
// that greatly reduce complexity of UserResponse struct definition: no need to separate UserIn
// (which doesn't use Optional) and UserOut (which uses Optional).
type Optional[T any] struct {
	Defined bool
	Value   *T
}

// Conventions followed: https://google.github.io/styleguide/jsoncstyleguide.xml#Reserved_Property_Names_in_the_error_object
type ErrorHandler func(http.ResponseWriter, *http.Request) (error, ErrorResponse)

type AppError struct {
	Err  error
	Resp ErrorResponse
}

type ErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message,omitempty"`
	Errors  []ErrorDescriptor `json:"errors,omitempty"`
}

type ErrorDescriptor struct {
	Domain       string `json:"domain,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Location     string `json:"location,omitempty"`
	LocationType string `json:"location,omitempty"`
}

type JsonDateTime time.Time

type JsonDuration struct {
	Duration int
	Unit     string
}

func (dt *JsonDateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(*dt)
	// convert to JSON string type
	return []byte(`"` + t.Format(DatetimeFormat) + `"`), nil
}

func (dt *JsonDateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := time.Parse(DatetimeFormat, string(b[1:len(b)-1]))
	if err != nil {
		return err
	}
	*dt = JsonDateTime(t)
	return nil
}

func (d *JsonDuration) MarshalJSON() ([]byte, error) {
	if d.Duration >= 100000 {
		return nil, errors.New("duration must be <= 100000")
	}
	if d.Unit != "sec" && d.Unit != "min" && d.Unit != "hour" {
		return nil, errors.New("invalid duration unit " + d.Unit)
	}
	return []byte(strconv.Itoa(d.Duration) + " " + d.Unit), nil
}

func (d *JsonDuration) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var dur int
	var unit string
	for i, c := range b {
		if c >= '0' && c <= '9' {
			dur = dur*10 + int(c)
			if dur >= 100000 {
				return errors.New("duration must be <= 100000")
			}
		}
		if c >= 'a' && c <= 'z' {
			unit = string(b[i:])
			break
		}
	}
	if unit != "sec" && unit != "min" && unit != "hour" {
		return errors.New("invalid duration unit " + unit)
	}
	*d = JsonDuration{
		Duration: dur,
		Unit:     unit,
	}
	return nil
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

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
	Db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled",
		Ev.Var(SqlHostVar),
		Ev.Var(SqlPortVar),
		Ev.Var(SqlUsernameVar),
		Ev.Var(SqlPasswordVar),
		Ev.Var(SqlWebDbNameVar)))

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
