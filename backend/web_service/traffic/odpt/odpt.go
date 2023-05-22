package odpt

import "net/http"

type odpt struct {
	Cl *http.Client

	ApiKey string
}
