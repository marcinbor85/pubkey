package user

import (
	"database/sql"
	"encoding/base64"
	"math/rand"
	"strings"
	"errors"
	"reflect"
	"time"

	"github.com/marcinbor85/pubkey/log"
)

type User struct {
	Id             	int64
	Username       	string
	Email          	string
	PublicKeyEncode	string
	PublicKeySign	string
	Active         	bool
	Deleted        	bool
	ActivateToken  	string
	DeleteToken    	string
	CreateDatetime 	time.Time
}

func UnpackStruct(u interface{}) []interface{} {
    val := reflect.ValueOf(u).Elem()
    v := make([]interface{}, val.NumField())
    for i := 0; i < val.NumField(); i++ {
        valueField := val.Field(i)
        v[i] = valueField.Addr().Interface()
    }
    return v
}

func CreateTable(db *sql.DB) error {
	sqlText := `CREATE TABLE TblUser (
		"Id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"Username" TEXT,
		"Email" TEXT,
		"PublicKeyEncode" TEXT,
		"PublicKeySign" TEXT,
		"Active" INTEGER DEFAULT 0,
		"Deleted" INTEGER DEFAULT 0,
		"ActivateToken" TEXT,
		"DeleteToken" TEXT,
		"CreateDatetime" DATETIME DEFAULT (DATETIME('NOW'))
	);`

	st, err := db.Prepare(sqlText)
	if err != nil {
		log.E(err.Error())
		return err
	}
	_, err = st.Exec()
	if err != nil {
		log.E(err.Error())
		return err
	}
	return nil
}

func Add(db *sql.DB, username string, email string, publickeyEncode string, publickeySign string) (*User, error) {
	username = strings.ToLower(username)
	email = strings.ToLower(email)

	user, err := GetByUsername(db, username)
	if user != nil {
		return nil, errors.New("user already exist")
	}

	sqlText := `INSERT INTO TblUser (Username, Email, PublicKeyEncode, PublicKeySign, ActivateToken, DeleteToken) VALUES (?,?,?,?,?,?)`
	st, err := db.Prepare(sqlText)
	if err != nil {
		log.E(err.Error())
		return nil, err
	}

	tokenArray := make([]byte, 32)
	rand.Read(tokenArray)
	actToken := base64.URLEncoding.EncodeToString(tokenArray)
	rand.Read(tokenArray)
	delToken := base64.URLEncoding.EncodeToString(tokenArray)

	_, err = st.Exec(username, email, publickeyEncode, publickeySign, actToken, delToken)
	if err != nil {
		log.E(err.Error())
		return nil, err
	}

	user, _ = GetByUsername(db, username)
	return user, nil
}

func GetByUsername(db *sql.DB, username string) (*User, error) {
	username = strings.ToLower(username)

	sqlText := `SELECT * FROM TblUser WHERE Deleted = 0 AND Username = ?`
	st, err := db.Prepare(sqlText)
	if err != nil {
		log.E(err.Error())
		return nil, err
	}

	user := User{}
	fieldPointers := UnpackStruct(&user)
	err = st.QueryRow(username).Scan(fieldPointers...)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func Activate(db *sql.DB, username string) error {
	username = strings.ToLower(username)

	sqlText := `UPDATE TblUser SET Active = 1 WHERE Username = ? AND Deleted = 0`
	st, err := db.Prepare(sqlText)

	if err != nil {
		log.E(err.Error())
		return err
	}

	_, err = st.Exec(username)
	if err != nil {
		log.E(err.Error())
		return err
	}

	return nil
}

func Delete(db *sql.DB, username string) error {
	username = strings.ToLower(username)

	sqlText := `UPDATE TblUser SET Deleted = 1 WHERE Username = ? AND Deleted = 0`
	st, err := db.Prepare(sqlText)

	if err != nil {
		log.E(err.Error())
		return err
	}

	_, err = st.Exec(username)
	if err != nil {
		log.E(err.Error())
		return err
	}

	return nil
}
