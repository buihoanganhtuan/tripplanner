package users

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
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
	// Fetch all necessary environment variables
	ev := &utils.EnvironmentVariableMap{}
	ev.Fetch("PQ_USERNAME", "PQ_PASSWORD", "PQ_AUTH_DBNAME", "PQ_AUTH_TABLENAME")
	ev.Fetch("SENDER_GMAIL_ACCOUNT", "SENDER_GMAIL_PASSWORD")
	if ev.Err() != nil {
		return fmt.Errorf("environmental variable error: %v", ev.Err()), http.StatusInternalServerError
	}

	// Parse HTTP Form contained in Body
	err := req.ParseForm()
	if err != nil {
		return fmt.Errorf("fail to parse request body: %v", err), http.StatusBadRequest
	}

	// Username and password sanity check
	if !req.Form.Has("email") || !req.Form.Has("password") || !req.Form.Has("username") {
		return fmt.Errorf("request form contains no email or password"), http.StatusBadRequest
	}
	var (
		email    = req.Form["email"][0]
		password = req.Form["password"][0]
		uname    = req.Form["username"][0]
	)

	if !utils.CheckPasswordStrength(password) || !utils.CheckEmailFormat(email) || !utils.CheckUsername(uname) {
		return fmt.Errorf("invalid email: %v or username: %v or password: %v", email, uname, password), http.StatusBadRequest
	}

	// User identity conflict check
	db, err := sql.Open("postgres", fmt.Sprintf("username=%s password=%s dbname=%s sslmode=disabled",
		ev.Var("PQ_USERNAME"),
		ev.Var("PQ_PASSWORD"),
		ev.Var("PQ_AUTH_DBNAME")))
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("database connection error: %v", err), http.StatusInternalServerError
	}

	rows, err := db.Query("SELECT email, verified, token_expiration FROM ? WHERE username = ? OR email = ? AS count",
		ev.Var("PQ_AUTH_TABLENAME"),
		uname,
		email)
	const datetimeFormat = "2020-12-24 23:59:00"
	for rows.Next() {
		var ver bool
		var exp, em string
		rows.Scan(&em, &ver, &exp)
		t, err := time.Parse(datetimeFormat, exp)
		if err != nil {
			return fmt.Errorf("cannot parse token expiration date %v for user %v", exp, email), http.StatusInternalServerError
		}
		if ver || t.Before(time.Now()) {
			if em == email {
				return fmt.Errorf("email %v already exist", email), http.StatusBadRequest
			}
			return fmt.Errorf("username %v already exist", uname), http.StatusBadRequest
		}
	}

	// Hash password and insert hashed password to database
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("fail to generate hash for password %v", err), http.StatusInternalServerError
	}

	rand.Seed(time.Now().Unix())
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, 24)
	for i := range token {
		token[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	now := time.Now().In(time.UTC)
	_, err = db.Exec("INSERT INTO ?(email, password, joined_date, validation_token, token_expiration, verified) VALUES(?, ?, ?, ?, ?, ?)",
		ev.Var("PQ_AUTH_TABLENAME"),
		email,
		string(hash),
		now.Format(datetimeFormat),
		string(token),
		now.Add(time.Duration(24*3600*1_000_000_000)).Format(datetimeFormat),
		"False")
	if err != nil {
		return fmt.Errorf("fail to insert new user into database: %v", err), http.StatusInternalServerError
	}

	// Send confirmation email to user and redirect user to /users/confirmation/username=&token=
	from := ev.Var("SENDER_GMAIL_ACCOUNT")
	pass := ev.Var("SENDER_GMAIL_PASSWORD")
	to := []string{email}
	sub := "Account confirmation at TripPlanner"
	msg := fmt.Sprintf(`Thank you for signing up to Trip Planner.
	Please click on the URL below to verify your account
	
	http://localhost/users/confirmation/username=%s&token=%s
	
	This URL is valid for 24 hours.`, uname, token)

	headers := map[string]string{
		"From":    from,
		"To":      to[0],
		"Subject": sub,
	}
	body := ""
	for k, v := range headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + msg

	auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	err = smtp.SendMail("smtp.gmail.com", auth, from, to, []byte(body))
	if err != nil {
		return fmt.Errorf("error sending confirmation email: %v", err), http.StatusInternalServerError
	}

	w.Header().Set("Location", fmt.Sprintf("http://localhost/confirmation?user=%s", uname))
	w.WriteHeader(http.StatusSeeOther)

	return nil, 0
}
