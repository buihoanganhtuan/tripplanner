package utils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// unsafe manager type to manage environment variables
// Attempting to get a variable via Var() when the manager is in error state or
// without fetching the variable first via Fetch() will result in a panic
type EnvironmentVariableMap struct {
	varMap map[string]string
	err    error
}





// Tagging to assist JSON Marshalling (converting a structured data into a JSON string)
// Note that in this case, missing, null, and zero values are handled in the same way so
// that greatly reduce complexity of UserResponse struct definition: no need to separate UserIn
// (which doesn't use Optional) and UserOut (which uses Optional).
type Optional[T any] struct {
	Defined bool
	Value   T
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

func (dt JsonDateTime) Compare(odt JsonDateTime) int {
	t1 := time.Time(dt)
	t2 := time.Time(odt)
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
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
	switch o.Value.(type) {
	case json.Unmarshaler:
		return o.Value.UnmarshalJSON(data)
	default:
		return json.Unmarshal(data, &o.Value)
	}
}






func (env *EnvironmentVariableMap) Fetch(names ...string) {
	if env.err != nil {
		return
	}
	if len(env.varMap) == 0 {
		env.varMap = make(map[string]string)
	}
	for _, name := range names {
		if _, exist := env.varMap[name]; exist {
			continue
		}
		if val, ok := os.LookupEnv(name); ok {
			env.varMap[name] = val
			continue
		}
		env.err = fmt.Errorf("environment variable %v is unset", name)
	}
}

func (env *EnvironmentVariableMap) Var(name string) string {
	if _, exist := env.varMap[name]; env.err != nil || !exist {
		if env.err != nil {
			panic(env.err)
		}
		panic(fmt.Errorf("environmental variable %v has not been fetched", name))
	}
	return env.varMap[name]
}

func GetBase32RandomString(length int) string {
	const b32Charset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	const checksumCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	b := make([]byte, length+1)
	checkSum := 0
	for i := 0; i < length; i++ {
		b[i] = b32Charset[rnd.Intn(32)]
		checkSum = (checkSum*(int('Z')+1) + int(b[i])) % 37
	}
	b[length] = checksumCharset[checkSum]
	return string(b)
}

func VerifyBase32String(s string) bool {
	s = strings.ToUpper(s)
	const checksumCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ*~$=U"
	checkSum := 0
	for i := 0; i < len(s)-1; i++ {
		checkSum = (checkSum*(int('Z')+1) + int(s[i])) % 37
	}
	return s[len(s)-1] == checksumCharset[checkSum]
}

func (env *EnvironmentVariableMap) Err() error {
	return env.err
}

func ExtractClaims(rq *http.Request, pk *rsa.PublicKey) (jwt.MapClaims, error) {
	// check if there is any access token
	if rq.Header.Get("Authorization") == "" {
		return nil, fmt.Errorf("no access token")
	}

	// check access token integrity. Note that we don't support BasicAuth
	ts, ok := strings.CutPrefix(rq.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(ts, func(token *jwt.Token) (interface{}, error) {
		return pk, nil
	}, jwt.WithValidMethods([]string{"RSA"}))

	if err != nil || !token.Valid {
		return nil, err
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claim datatype")
	}

	return mc, nil
}

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var ret []K
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	var ret []V
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}

func NewInvalidIdError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "invalid resource id",
	}
}

func NewServerParseError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "server-side parsing error",
	}
}

func NewClientParseError(field string) ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("fail to parse field %s", field),
	}
}

func NewDatabaseQueryError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "database query error",
	}
}

func NewDatabaseTransactionError() ErrorResponse {
	return ErrorResponse{
		Code: http.StatusInternalServerError,
		Message: "database transaction error"
	}
}

func NewUnknownError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "unknown server-side error",
	}
}

func NewMarshalError() ErrorResponse {
	return ErrorResponse{
		Code: http.StatusInternalServerError,
		Message: "fail to marshal response body",
	}
}

func NewUnmarshalError() ErrorResponse {
	return ErrorResponse{
		Code: http.StatusBadRequest,
		Message: "fail to unmarshal request body",
	}
}