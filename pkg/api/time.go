package api

import (
	"database/sql"
	"time"
)

func now() *sql.Null[time.Time] {
	t := time.Now().UTC()
	return &sql.Null[time.Time]{V: t, Valid: true}
}
