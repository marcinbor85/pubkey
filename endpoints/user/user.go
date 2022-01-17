package user

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/google/uuid"

	"github.com/marcinbor85/pubkey/config"
	"github.com/marcinbor85/pubkey/crypto"
	"github.com/marcinbor85/pubkey/crypto/rsa"
	"github.com/marcinbor85/pubkey/database"
	"github.com/marcinbor85/pubkey/email"
	"github.com/marcinbor85/pubkey/log"

	mUser "github.com/marcinbor85/pubkey/models/user"

	"text/template"

	"regexp"
)

const USERNAME_REGEX string = `[a-z0-9_-]{3,32}`
const ENDPOINT_NAME string = "users"

const ACTIVATE_TOKEN_EXPIRE_DURATION = 24*time.Hour

func validateUsername(username string) bool {
	rule := "^" + USERNAME_REGEX + "$"
	re := regexp.MustCompile(rule)
	return re.MatchString(username)
}

func validateEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	user := parts[0]
	if len(user) < 1 {
		return false
	}
	domain := parts[1]
	if len(domain) < 1 {
		return false
	}
	if len(email) > 32 {
		return false
	}
	return true
}

func Register(router *mux.Router) {
	router.HandleFunc("/"+ENDPOINT_NAME, addEndpoint).Methods(http.MethodPost)

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

	log.D("request addEndpoint")

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

	ok = validateEmail(u.Email)
	if ok == false {
		http.Error(w, "email validation error", http.StatusBadRequest)
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

	log.I("user added: %s", user.Username)

	client := &email.Client{
		Email:    config.Get("SMTP_EMAIL"),
		Password: config.Get("SMTP_PASSWORD"),
		Host:     config.Get("SMTP_HOST"),
		Port:     config.Get("SMTP_PORT"),
	}

	type EmailContext struct {
		Username     string
		APILink      string
		ActivateLink string
		DeleteLink   string
	}

	tmpl, err := template.ParseFiles(config.Get("TEMPLATE_WELCOME_EMAIL"))
	if err != nil {
		log.E("parsing email template: %s", err.Error())
	}

	host := config.Get("HOST")

	context := EmailContext{
		Username:     user.Username,
		APILink:      strings.Join([]string{host, ENDPOINT_NAME, user.Username}, "/"),
		ActivateLink: strings.Join([]string{host, ENDPOINT_NAME, user.Username, user.ActivateToken}, "/"),
		DeleteLink:   strings.Join([]string{host, ENDPOINT_NAME, user.Username, user.DeleteToken}, "/"),
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

	err = client.Send(em)
	if err != nil {
		log.E("sending email error: %s", err.Error())
	}

	w.WriteHeader(http.StatusCreated)
}

func getEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username := params["username"]

	log.D("request getEndpoint: %s", username)

	user, err := mUser.GetByUsername(database.DB, username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if user.Active == false {
		http.NotFound(w, r)
		return
	}

	type Data struct {
		PublicKey		string `json:"public_key"`
		Uuid		 	string `json:"uuid"`
	}
	
	data := &Data{
		PublicKey: user.PublicKey,
		Uuid: uuid.New().String(),
	}
	text, err := json.Marshal(data)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	dataString := string(text)
	
	privateKey := config.G.PrivateKey
	dataBin := []byte(dataString)
	signature, err := rsa.Sign(dataBin, privateKey)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	signatureEncoded := base64.URLEncoding.EncodeToString(signature)

	w.Header().Set("Content-Type", "application/json")

	type Response struct {
		Data		*Data `json:"response"`
		Signature	string `json:"signature"`
	}
	response := &Response{
		Data: data,
		Signature: signatureEncoded,
	}

	enc := json.NewEncoder(w)
	enc.Encode(response)
}

func tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	username := params["username"]
	token := params["token"]

	log.D("request tokenEndpoint: %s", username)

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
			createTime := user.CreateDatetime
			expireTime := createTime.Add(time.Duration(ACTIVATE_TOKEN_EXPIRE_DURATION))
			if time.Now().UTC().After(expireTime) {
				http.Error(w, "activate token expired", http.StatusGone)
				return
			}

			err = mUser.Activate(database.DB, username)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			log.I("user activated: %s", username)
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

			log.I("user deleted: %s", username)
			fmt.Fprintln(w, "user deleted")
		}
	}
}
