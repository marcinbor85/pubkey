package user

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"math/rand"
)

type User struct {
	Id            int64
	Username      string
	Email         string
	PublicKey     string
	Active        bool
	Deleted       bool
	ActivateToken string
	DeleteToken   string
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
		"DeleteToken" TEXT
	);`

	st, err := db.Prepare(sqlText)
	if err != nil {
		return err
	}
	_, err = st.Exec()
	if err != nil {
		return err
	}
	return nil
}

func Add(db *sql.DB, username string, email string, publickey string) (*User, error) {
	u, _ := GetByUsername(db, username)
	if u != nil {
		return nil, errors.New("user already exist")
	}

	sqlText := `INSERT INTO TblUser (Username, Email, PublicKey, ActivateToken, DeleteToken) VALUES (?,?,?,?,?)`
	st, err := db.Prepare(sqlText)
	if err != nil {
		return nil, err
	}

	tokenArray := make([]byte, 32)
	rand.Read(tokenArray)
	actToken := base64.URLEncoding.EncodeToString(tokenArray)
	rand.Read(tokenArray)
	delToken := base64.URLEncoding.EncodeToString(tokenArray)

	user := &User{
		Id:            -1,
		Username:      username,
		Email:         email,
		PublicKey:     publickey,
		Active:        false,
		Deleted:       false,
		ActivateToken: actToken,
		DeleteToken:   delToken,
	}

	res, err := st.Exec(user.Username, user.Email, user.PublicKey, user.ActivateToken, user.DeleteToken)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.Id = id
	return user, nil
}

func GetByUsername(db *sql.DB, username string) (*User, error) {
	sqlText := `SELECT * FROM TblUser WHERE Deleted = 0 AND Username = ?`
	st, err := db.Prepare(sqlText)

	if err != nil {
		return nil, err
	}

	user := &User{}
	err = st.QueryRow(username).Scan(&user.Id, &user.Username, &user.Email, &user.PublicKey, &user.Active, &user.Deleted, &user.ActivateToken, &user.DeleteToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func Activate(db *sql.DB, username string) error {
	sqlText := `UPDATE TblUser SET Active = 1 WHERE Username = ? AND Deleted = 0`
	st, err := db.Prepare(sqlText)

	if err != nil {
		return err
	}

	_, err = st.Exec(username)
	if err != nil {
		return err
	}

	return nil
}

func Delete(db *sql.DB, username string) error {
	sqlText := `UPDATE TblUser SET Deleted = 1 WHERE Username = ? AND Deleted = 0`
	st, err := db.Prepare(sqlText)

	if err != nil {
		return err
	}

	_, err = st.Exec(username)
	if err != nil {
		return err
	}

	return nil
}
