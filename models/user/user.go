package user

import (
	"database/sql"
	"encoding/base64"
	"math/rand"
	"errors"
	"time"

	"github.com/marcinbor85/pubkey/log"
)

type User struct {
	Id             int64
	Username       string
	Email          string
	PublicKey      string
	Active         bool
	Deleted        bool
	ActivateToken  string
	DeleteToken    string
	CreateDatetime time.Time
}

func CreateTable(db *sql.DB) error {
	sqlText := `CREATE TABLE TblUser (
		"Id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"Username" TEXT,
		"Email" TEXT,
		"PublicKey" TEXT,
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

func Add(db *sql.DB, username string, email string, publickey string) (*User, error) {
	user, err := GetByUsername(db, username)
	if user != nil {
		return nil, errors.New("user already exist")
	}

	sqlText := `INSERT INTO TblUser (Username, Email, PublicKey, ActivateToken, DeleteToken) VALUES (?,?,?,?,?)`
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

	_, err = st.Exec(username, email, publickey, actToken, delToken)
	if err != nil {
		log.E(err.Error())
		return nil, err
	}

	user, _ = GetByUsername(db, username)
	return user, nil
}

func GetByUsername(db *sql.DB, username string) (*User, error) {
	sqlText := `SELECT * FROM TblUser WHERE Deleted = 0 AND Username = ?`
	st, err := db.Prepare(sqlText)
	if err != nil {
		log.E(err.Error())
		return nil, err
	}

	var createDatetime string

	user := &User{}
	err = st.QueryRow(username).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.PublicKey,
		&user.Active,
		&user.Deleted,
		&user.ActivateToken,
		&user.DeleteToken,
		&createDatetime,
	)
	if err != nil {
		return nil, err
	}
	parsedCreateDatetime, e := time.Parse(time.RFC3339, createDatetime)
	if e != nil {
		log.E(e.Error())
		return nil, e
	}
	user.CreateDatetime = parsedCreateDatetime
	return user, nil
}

func Activate(db *sql.DB, username string) error {
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
