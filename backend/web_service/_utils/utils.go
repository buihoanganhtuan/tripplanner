package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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

type jwtChecker struct {
	mapClaims jwt.MapClaims
	errClaim  string
}

func (c *jwtChecker) checkClaim(name string, val interface{}, req bool) {
	if c.errClaim != "" {
		return
	}
	switch name {
	case "iss":
		sval, ok := val.(string)
		if !ok || !c.mapClaims.VerifyIssuer(sval, req) {
			c.errClaim = name
		}
	case "sub":
		sval, _ := val.(string)
		mval, _ := c.mapClaims[name].(string)
		if sval == "" || mval == "" {
			c.errClaim = name
		}
	case "iat":
		ival, ok := val.(int64)
		if !ok || !c.mapClaims.VerifyIssuedAt(ival, req) {
			c.errClaim = name
		}
	case "exp":
		ival, ok := val.(int64)
		if !ok || !c.mapClaims.VerifyExpiresAt(ival, req) {
			c.errClaim = name
		}
	default:
		v, present := c.mapClaims["name"]
		if present && v != val || !present && req {
			c.errClaim = name
		}
	}
}

func (c *jwtChecker) Err() error {
	if c.errClaim != "" {
		return fmt.Errorf("error for claim %s", c.errClaim)
	}
	return nil
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

func VerifyBase32String(s string, length int) bool {
	if len(s) != length+1 {
		return false
	}
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
