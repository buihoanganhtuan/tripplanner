package users

import (
	"fmt"

	jwt "github.com/golang-jwt/jwt/v4"
)

// var fileServer = http.StripPrefix("/users", http.FileServer(http.Dir("./users/assets/")))

// ********************** Auxiliary types *******************************
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

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
