package constants

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/golang-jwt/jwt/v4"
)

var Database *sql.DB
var PublicKey *rsa.PublicKey
var EnvironmentVariable utils.EnvironmentVariableMap

const (
	DATE_TIME_FORMAT    = "2006-01-02 15:04:05 -0700"
	PQ_HOST_VAR         = "PQ_HOST"
	PQ_PORT_VAR         = "PQ_PORT"
	PQ_USERNAME_VAR     = "PQ_USERNAME"
	PQ_PASSWORD_VAR     = "PQ_PASSWORD"
	PQ_WEB_DBNAME_VAR   = "PQ_WEB_DBNAME"
	PQ_USER_TABLE_VAR   = "PQ_USER_TABLE"
	PUBLIC_KEY_PATH_VAR = "PUBLIC_KEY_PATH"
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
		PQ_HOST_VAR,
		PQ_PORT_VAR,
		PQ_USERNAME_VAR,
		PQ_PASSWORD_VAR,
		PQ_WEB_DBNAME_VAR,
		PQ_USER_TABLE_VAR,
		PUBLIC_KEY_PATH_VAR)

	var err error
	Database, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s username=%s password=%s dbname=%s sslmode=disabled",
		EnvironmentVariable.Var(PQ_HOST_VAR),
		EnvironmentVariable.Var(PQ_PORT_VAR),
		EnvironmentVariable.Var(PQ_USERNAME_VAR),
		EnvironmentVariable.Var(PQ_PASSWORD_VAR),
		EnvironmentVariable.Var(PQ_WEB_DBNAME_VAR)))

	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}
	err = Database.Ping()
	if err != nil {
		panic(fmt.Errorf("database connection error: %v", err))
	}

	// Load authentication server's public key for access token validation
	EnvironmentVariable.Fetch(PUBLIC_KEY_PATH_VAR, PQ_USER_TABLE_VAR)
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
}
