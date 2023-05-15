package rest

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/buihoanganhtuan/tripplanner/backend/web_service/domain"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/encoding/base32"
	"github.com/buihoanganhtuan/tripplanner/backend/web_service/environment/variables"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

const (
	publicKeyPathVar            = "PUBLIC_KEY_PATH"
	authServiceName             = "Tripplanner:AuthService"
	webServiceName              = "Tripplanner:WebService"
	unauthorizedMsg             = "user is unauthorized"
	unauthorizedNonTokenMsg     = "only JWT token allowed"
	unauthorizedInvalidTokenMsg = "invalid JWT token"
	unauthorizedInvalidClaimMsg = "invalid claim %s"
)

type Rest struct {
	Dom  domain.Domain
	Serv http.Server
	ev   variables.EnvironmentVariableMap
	pk   *rsa.PublicKey
}

func (r *Rest) Init() {
	// Load authentication server's public key for access token validation
	r.ev.Fetch(publicKeyPathVar)
	if r.ev.Err() != nil {
		panic(fmt.Errorf("environment variable error: %v", Ev.Err()))
	}

	b, err := os.ReadFile(r.ev.Var(publicKeyPathVar))
	if err != nil {
		panic(fmt.Errorf("cannot read public key file: %v", err))
	}

	r.pk, err = jwt.ParseRSAPublicKeyFromPEM(b)
	if err != nil {
		panic(fmt.Errorf("fail to parse public key from file: %v", err))
	}
}

func (r *Rest) GetUser(id domain.UserId) (domain.User, error) {

}

// For consistency, we should select a convention for the response when errors occur and stick with it.
// The convention chosen here is: https://google.github.io/styleguide/jsoncstyleguide.xml#Reserved_Property_Names_in_the_error_object
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
	LocationType string `json:"locationType,omitempty"`
}

type ErrorHandler func(w http.ResponseWriter, r *http.Request) (ErrorResponse, error)
type Middleware func(ErrorHandler) http.HandlerFunc
type AclClaim struct {
	jwt.StandardClaims
	User        string   `json:"user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"perms"`
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
					if base32.Verify(tokens[i]) {
						continue
					}

					er := ErrorResponse{
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

				token, err := jwt.ParseWithClaims(authHead, &AclClaim{}, func(token *jwt.Token) (interface{}, error) {
					return Pk, nil
				}, jwt.WithValidMethods([]string{"RSA"}))

				claims, ok := token.Claims.(AclClaim)
				if err != nil || !token.Valid || !ok {
					SimpleUnauthorizeResponse(w, unauthorizedInvalidTokenMsg)
					return
				}

				if !claims.VerifyIssuer(authServiceName, true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "iss"))
					return
				}

				if !claims.VerifyAudience(webServiceName, true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "aud"))
					return
				}

				if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "exp"))
					return
				}

				// check permission
				if len(claims.Permissions) == 0 {
					SimpleUnauthorizeResponse(w, fmt.Sprintf(unauthorizedInvalidClaimMsg, "perms"))
				}

				method := strings.ToLower(r.Method)
				if method == "post" && strings.ContainsRune(method, ':') {
					method = strings.ToLower(peekBack(strings.Split(method, ":")))
				}
				ok = false
				for _, p := range claims.Permissions {
					p = strings.ToLower(p)
					rm := strings.Split(p, ":")
					ok = ok || len(rm) == 2 && rm[1] == method && matchResource(r.URL.Path, rm[0])
				}
				if !ok {
					SimpleUnauthorizeResponse(w, unauthorizedMsg)
				}
			}

			// call the inner handler
			er, e := h(w, r)
			if e != nil {
				fmt.Printf("%v", e)
				resp, err := json.Marshal(er)
				if err != nil {
					panic(err)
				}
				http.Error(w, string(resp), er.Code)
			}
		})
	}
}

func matchResource(path, resource string) bool {
	// TODO: implement a custom JSON-valued claim to implement authorization
	// We have a set of (resource, method) pairs. Each security role refers to
	// a specific set of those pairs and indicates that the role can perform
	// those specific methods on those specific resources. An identity can assume
	// one or multiple roles, depending on our policy
	path = strings.ToLower(path)
	resource = strings.ToLower(resource)

	// matching
	if resource[len(resource)-1] == '*' {
		return strings.HasPrefix(path, resource)
	}
	return path == resource
}

func SimpleUnauthorizeResponse(w http.ResponseWriter, msg string) {
	resp, err := json.Marshal(ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: msg,
	})
	if err != nil {
		panic(err)
	}
	http.Error(w, string(resp), http.StatusUnauthorized)
}

func peekBack[T any](arr []T) T {
	var ret T
	if len(arr) > 0 {
		ret = arr[len(arr)-1]
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
		Code:    http.StatusInternalServerError,
		Message: "database transaction error",
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
		Code:    http.StatusInternalServerError,
		Message: "fail to marshal response body",
	}
}

func NewUnmarshalError() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "fail to unmarshal request body",
	}
}
