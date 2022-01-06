package user

import (
	"fmt"
	"net/http"
	"strings"
	"bytes"

	"encoding/json"
	"github.com/gorilla/mux"

	"github.com/marcinbor85/pubkey/config"
	"github.com/marcinbor85/pubkey/crypto"
	"github.com/marcinbor85/pubkey/log"
	"github.com/marcinbor85/pubkey/database"
	"github.com/marcinbor85/pubkey/email"

	mUser "github.com/marcinbor85/pubkey/models/user"

	"text/template"

	"regexp"
)

const USERNAME_REGEX string = `[a-zA-Z0-9_-]{3,}`
const ENDPOINT_NAME string = "user"

func validateUsername(username string) bool {
	rule := "^" + USERNAME_REGEX + "$"
	re := regexp.MustCompile(rule)
	return re.MatchString(username)
}

func Register(router *mux.Router) {
	router.HandleFunc("/" + ENDPOINT_NAME, addEndpoint).Methods(http.MethodPost)

	path := "/" + ENDPOINT_NAME + "/" + "{username:" + USERNAME_REGEX + "}"
	router.HandleFunc(path, getEndpoint).Methods(http.MethodGet)

	path = "/" + ENDPOINT_NAME + "/" + "{username:" + USERNAME_REGEX + "}/{token}"
	router.HandleFunc(path, tokenEndpoint).Methods(http.MethodGet)
}

func addEndpoint(w http.ResponseWriter, r *http.Request) {
	type RegisterUser struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		PublicKey string `json:"public_key"`
	}

	var u RegisterUser

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := validateUsername(u.Username)
	if ok == false {
		http.Error(w, "username validation error", http.StatusBadRequest)
		return
	}

	_, err = crypto.DecodePublicKey(u.PublicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := mUser.Add(database.DB, u.Username, u.Email, u.PublicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := &email.Client{
		Email:    config.Get("SMTP_EMAIL"),
		Password: config.Get("SMTP_PASSWORD"),
		Host:     config.Get("SMTP_HOST"),
		Port:     config.Get("SMTP_PORT"),
	}

	type EmailContext struct {
		Username string
		APILink string
		ActivateLink string
		DeleteLink string
	}
	
	tmpl, err := template.ParseFiles(config.Get("TEMPLATE_WELCOME_EMAIL"))
	if err != nil {
		log.E("parsing email template: %s", err.Error())
	}
	
	host := config.Get("HOST")

	context := EmailContext{
		Username: user.Username,
		APILink: strings.Join([]string{host, ENDPOINT_NAME, user.Username}, "/"),
		ActivateLink: strings.Join([]string{host, ENDPOINT_NAME, user.Username, user.ActivateToken}, "/"),
		DeleteLink: strings.Join([]string{host, ENDPOINT_NAME, user.Username, user.DeleteToken}, "/"),
	}

	var msgBuffer bytes.Buffer
	err = tmpl.Execute(&msgBuffer, context)
	if err != nil {
		log.E("template executing: %s", err.Error())
	}

	em := &email.Email{
		ToName:    u.Username,
		ToEmail:   u.Email,
		FromName:  config.Get("TEXT_EMAIL_SENDER"),
		FromEmail: config.Get("SMTP_EMAIL"),
		Subject:   config.Get("TEXT_WELCOME_EMAIL_SUBJECT"),
		Message:   msgBuffer.String(),
	}

	_ = client.Send(em)

	w.WriteHeader(http.StatusCreated)
}

func getEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username := params["username"]

	user, err := mUser.GetByUsername(database.DB, username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if user.Active == false {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	d := map[string]string{"username": user.Username, "public_key": user.PublicKey}
	enc.Encode(d)
}

func tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username := params["username"]
	token := params["token"]

	user, err := mUser.GetByUsername(database.DB, username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if user.Active == false {
		if user.ActivateToken != token {
			http.NotFound(w, r)
			return
		} else {
			err = mUser.Activate(database.DB, username)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			fmt.Fprintln(w, "user activated")
		}
	} else {
		if user.DeleteToken != token {
			http.NotFound(w, r)
			return
		} else {
			err = mUser.Delete(database.DB, username)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			fmt.Fprintln(w, "user deleted")
		}
	}
}
