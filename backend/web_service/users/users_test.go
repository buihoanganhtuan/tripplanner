package users

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func TestGetUser(t *testing.T) {

}

func TestListUsers(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/users", utils.ErrorHandler(ListUsers)).Methods("GET")

	go http.ListenAndServe("localhost:8080", r)

	time.Sleep(time.Duration(5))
	resp, err := http.Get("http://localhost:8080/users?joinDate=less(2020-12-12)&joinDate=more(2019-12-01)")
	if err != nil {
		t.Fatalf("error %v", err)
	}

	bd, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	t.Log(resp.StatusCode)
	t.Log(string(bd))
	t.Log(resp.Header.Get("Content-Type"))
}
