package tasks

import (
	"database/sql"

	"github.com/marcinbor85/pubkey/log"
)

func DeleteExpiredRows(db *sql.DB) error {
	log.D("DeleteExpiredRows")

	sqlText := `UPDATE TblUser SET Deleted = 1 WHERE Active = 0 AND Deleted = 0 AND DATETIME(CreateDatetime, '+24 hours') < DATETIME('now')`
	st, err := db.Prepare(sqlText)

	if err != nil {
		log.E(err.Error())
		return err
	}

	res, err := st.Exec()
	if err != nil {
		log.E(err.Error())
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		log.E(err.Error())
		return err
	}

	log.I("DeleteExpiredRows rows_affected = %v", n)
	return nil
}
