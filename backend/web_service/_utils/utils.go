package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
)

func ErrorHandler(f func(w http.ResponseWriter, rq *http.Request) (int, string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		code, msg, err := f(w, rq)
		if err != nil {
			w.WriteHeader(code)
			io.WriteString(w, msg)
			log.Println(err)
		}
	}
}

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

func (env *EnvironmentVariableMap) Err() error {
	return env.err
}

func ValidateAccessToken(rq *http.Request, pk *rsa.PublicKey) (*jwt.Token, error) {
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

	return token, nil
}

func CheckPasswordStrength(passwd string) bool {
	noUpper, noDigit, noSpecial := true, true, true
	for _, c := range passwd {
		if isUpper(c) {
			noUpper = false
		}
		if isDigit(c) {
			noDigit = false
		}
		if !(isUpper(c) || isLower(c) || isDigit(c)) {
			noSpecial = false
		}
	}

	return !(noUpper || noDigit || noSpecial || len(passwd) < 8)
}

func CheckEmailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func CheckUsername(username string) bool {
	for _, c := range username {
		if !isUpper(c) && !isLower(c) && !isDigit(c) {
			return false
		}
	}
	return len(username) > 0 && len(username) <= 30
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
