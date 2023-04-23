package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/users"
	"github.com/golang-jwt/jwt/v4"
	mux "github.com/gorilla/mux"
)

const (
	unauthorizedMsg             = "user is unauthorized"
	unauthorizedNonTokenMsg     = "only JWT token allowed"
	unauthorizedInvalidTokenMsg = "invalid JWT token"
	unauthorizedInvalidClaimMsg = "invalid claim %s"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/users", users.ErrorHandler(users.CreateUser)).Methods("POST")
	r.HandleFunc("/{resource.id=users/.+}/", users.ErrorHandler(users.UpdateUser)).Methods("PATCH")
	r.HandleFunc("/{resource.id=users/.+}/", users.ErrorHandler(users.ReplaceUser)).Methods("PUT")
	r.HandleFunc("/{id=users/.*}/", users.ErrorHandler(users.GetUser)).Methods("GET")
	r.HandleFunc("/users{query=\\?.+}", users.ErrorHandler(users.ListUsers)).Methods("GET")
	r.HandleFunc("/{id=users/.+}/", users.ErrorHandler(users.DeleteUser)).Methods("DELETE")

}

// ErrorHandler return a pointer to ErrorResponse to help distinguish null from zero value
type ErrorHandler func(http.ResponseWriter, *http.Request) (error, cst.ErrorResponse)
type Middleware func(ErrorHandler) http.HandlerFunc
type AclClaim struct {
	jwt.StandardClaims
	user        string
	roles       []string
	permissions [][]string
}

func newValidatorMiddleware(conf map[string]interface{}) Middleware {
	return func(h ErrorHandler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// validate resource id (if any)
			varMap := mux.Vars(r)
			for _, v := range []string{"id", "resource.id", "parent"} {
				id, present := varMap[v]
				if !present {
					continue
				}
				tokens := strings.Split(id, "/")
				for i := 1; i < len(tokens); i += 2 {
					if utils.VerifyBase32String(tokens[i]) {
						continue
					}

					er := cst.ErrorResponse{
						Code:    http.StatusBadRequest,
						Message: fmt.Sprintf("invalid %s id %s", tokens[i-1][:len(tokens[i-1])-1], tokens[i]),
					}
					resp, err := json.Marshal(er)
					if err != nil {
						panic(err)
					}
					http.Error(w, string(resp), er.Code)
					return
				}
			}

			// validate access token
			if r.URL.Path != "" && !strings.HasPrefix(r.URL.Path, "trips/") {
				authHead := r.Header.Get("authorization")
				if authHead == "" && r.URL.Path != "" && strings.HasPrefix(r.URL.Path, "trips/") {
					SimpleUnauthorizeResponse(w, unauthorizedMsg)
					return
				}
				var hasToken bool
				authHead, hasToken = strings.CutPrefix(authHead, "Bearer ")
				if !hasToken {
					SimpleUnauthorizeResponse(w, unauthorizedNonTokenMsg)
					return
				}

				token, err := jwt.ParseWithClaims(authHead, AclClaim{}, func(token *jwt.Token) (interface{}, error) {
					return cst.Pk, nil
				}, jwt.WithValidMethods([]string{"RSA"}))

				claims, ok := token.Claims.(AclClaim)
				if err != nil || !token.Valid || !ok {
					SimpleUnauthorizeResponse(w, unauthorizedInvalidTokenMsg)
					return
				}

				if !claims.VerifyIssuer(cst.AuthServiceName, true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "iss"))
					return
				}

				if !claims.VerifyAudience(cst.WebServiceName, true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "aud"))
					return
				}

				if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "exp"))
					return
				}

				// check permission
				method := strings.ToLower(r.Method)
				if method == "POST" && strings.ContainsRune(method, ':') {
					tmp := strings.Split(method, ":")
					method = strings.ToLower(tmp[len(tmp)-1])
				}

				resource := r.URL.Path
				claims.
			}

			// call the inner handler
			e, er := h(w, r)
			if e != nil {
				fmt.Errorf("%v", e)
				resp, err := json.Marshal(er)
				if err != nil {
					panic(err)
				}
				http.Error(w, string(resp), er.Code)
			}
		})
	}
}

func SimpleUnauthorizeResponse(w http.ResponseWriter, msg string) {
	resp, err := json.Marshal(cst.ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: msg,
	})
	if err != nil {
		panic(err)
	}
	http.Error(w, string(resp), http.StatusUnauthorized)
}
