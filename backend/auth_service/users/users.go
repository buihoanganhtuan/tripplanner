package users

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/buihoanganhtuan/tripplanner/backend/auth_service/utils"
)

func ErrorHandler(f func(w http.ResponseWriter, rq *http.Request) (error, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		err, statusCode := f(w, rq)
		if err != nil {
			w.WriteHeader(statusCode)
			log.Println(err)
		}
	}
}

func UsersPostHandler(w http.ResponseWriter, req *http.Request) (error, int) {
	// Parse HTTP Form contained in Body
	err := req.ParseForm()
	if err != nil {
		return fmt.Errorf("fail to parse request body: %v", err), http.StatusBadRequest
	}

	// Username and password sanity check
	if !req.Form.Has("email") || !req.Form.Has("password") {
		return fmt.Errorf("request form contains no email or password"), http.StatusBadRequest
	}
	userEmail, userPassword := req.Form["email"][0], req.Form["password"][0]
	if !utils.CheckPasswordStrength(userPassword) || !utils.CheckEmailFormat(userEmail) {
		return fmt.Errorf("invalid email: %v or password: %v"), http.StatusBadRequest
	}

	// User email conflict check
	ev := &utils.EnvironmentVariableMap{}
	ev.Fetch("PQ_USERNAME", "PQ_PASSWORD", "PQ_AUTH_DBNAME", "PQ_AUTH_TABLENAME")
	if ev.Err() != nil {
		return fmt.Errorf("environmental variable error: %v", ev.Err()), http.StatusInternalServerError
	}

	db, err := sql.Open("postgres", fmt.Sprintf("username=%s password=%s dbname=%s sslmode=disabled",
		ev.Var("PQ_USERNAME"),
		ev.Var("PQ_PASSWORD"),
		ev.Var("PQ_AUTH_DBNAME")))
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("database connection error: %v", err), http.StatusInternalServerError
	}

	rows, err := db.Query("SELECT COUNT(*) FROM ? WHERE email = ? AS count", ev.Var("PQ_AUTH_TABLENAME"), userEmail)
	for rows.Next() {
		var count int
		rows.Scan(&count)
		if count > 0 {
			return fmt.Errorf("email %v already exist", userEmail), http.StatusBadRequest
		}
	}

	// Hash password and insert hashed password to database
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("fail to generate hash for password %v", err), http.StatusInternalServerError
	}

	_, err = db.Exec("INSERT INTO ?(email, password, joined_date) VALUES(?, ?, ?)",
		ev.Var("PQ_AUTH_TABLENAME"),
		userEmail,
		string(hashedPasswordBytes),
		time.Now().In(time.UTC).Format("2020-12-24"))
	if err != nil {
		return fmt.Errorf("fail to insert new user into database: %v", err), http.StatusInternalServerError
	}

	// Send confirmation email to user and redirect user to /confirmation/username

}
