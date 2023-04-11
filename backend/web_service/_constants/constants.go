package constants

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

var Database *sql.DB
var KvStoreDelete *redis.Client
var PublicKey *rsa.PublicKey
var EnvironmentVariable utils.EnvironmentVariableMap

const (
	DATE_TIME_FORMAT             = "2006-01-02 15:04:05 -0700"
	SQL_HOST_VAR                 = "PQ_HOST"
	SQL_PORT_VAR                 = "PQ_PORT"
	SQL_USERNAME_VAR             = "PQ_USERNAME"
	SQL_PASSWORD_VAR             = "PQ_PASSWORD"
	SQL_WEB_DBNAME_VAR           = "PQ_WEB_DBNAME"
	SQL_USER_TABLE_VAR           = "PQ_USER_TABLE"
	SQL_TRIP_TABLE_VAR           = "PQ_TRIP_TABLE"
	KV_HOST_VAR                  = "REDIS_HOST"
	KV_PORT_VAR                  = "REDIS_PORT"
	KV_PASSWORD_VAR              = "REDIS_PASSWORD"
	KV_DELETE_TRANSACTION_DB_VAR = "REDIS_DELETE_TRANSACTION_DBNAME"
	PUBLIC_KEY_PATH_VAR          = "PUBLIC_KEY_PATH"
)

// Tagging to assist JSON Marshalling (converting a structured data into a JSON string)
// Note that in this case, missing, null, and zero values are handled in the same way so
// that greatly reduce complexity of UserResponse struct definition: no need to separate UserIn
// (which doesn't use Optional) and UserOut (which uses Optional).
type UserResponse struct {
	Id       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	JoinDate DateTime `json:"joinDate,omitempty"`
}

type UserRequest struct {
	Id   string           `json:"id"`
	Name Optional[string] `json:"name"`
}

type Optional[T any] struct {
	Defined bool
	Value   *T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

type DateTime struct {
	Year   string `json:"year"`
	Month  string `json:"month"`
	Day    string `json:"day"`
	Hour   string `json:"hour"`
	Min    string `json:"min"`
	Sec    string `json:"sec"`
	Offset string `json:"timezone"`
}

func init() {
	EnvironmentVariable.Fetch(
		SQL_HOST_VAR,
		SQL_PORT_VAR,
		SQL_USERNAME_VAR,
		SQL_PASSWORD_VAR,
		SQL_WEB_DBNAME_VAR,
		SQL_USER_TABLE_VAR,
		SQL_TRIP_TABLE_VAR,
		KV_HOST_VAR,
		KV_PORT_VAR,
		KV_PASSWORD_VAR,
		KV_DELETE_TRANSACTION_DB_VAR,
		PUBLIC_KEY_PATH_VAR)

	var err error
	Database, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled",
		EnvironmentVariable.Var(SQL_HOST_VAR),
		EnvironmentVariable.Var(SQL_PORT_VAR),
		EnvironmentVariable.Var(SQL_USERNAME_VAR),
		EnvironmentVariable.Var(SQL_PASSWORD_VAR),
		EnvironmentVariable.Var(SQL_WEB_DBNAME_VAR)))

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
	err = Database.Ping()
	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}

	// Load authentication server's public key for access token validation
	EnvironmentVariable.Fetch(PUBLIC_KEY_PATH_VAR, SQL_USER_TABLE_VAR)
	if EnvironmentVariable.Err() != nil {
		panic(fmt.Errorf("environment variable error: %v", EnvironmentVariable.Err()))
	}

	b, err := os.ReadFile(EnvironmentVariable.Var(PUBLIC_KEY_PATH_VAR))
	if err != nil {
		panic(fmt.Errorf("cannot read public key file: %v", err))
	}

	PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		panic(fmt.Errorf("fail to parse public key from file: %v", err))
	}

	// Connect to key-value store
	dbn, err := strconv.Atoi(EnvironmentVariable.Var(KV_DELETE_TRANSACTION_DB_VAR))
	if err != nil {
		panic(fmt.Errorf("cannot parse %v as an int", EnvironmentVariable.Var(KV_DELETE_TRANSACTION_DB_VAR)))
	}
	KvStore := redis.NewClient(&redis.Options{
		Addr:     EnvironmentVariable.Var(KV_HOST_VAR) + ":" + EnvironmentVariable.Var(KV_PORT_VAR),
		Password: EnvironmentVariable.Var(KV_PASSWORD_VAR),
		DB:       dbn,
	})

	ctx := context.Background()
	_, err = KvStore.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("error connecting to key-value store: %v", err))
	}
}
