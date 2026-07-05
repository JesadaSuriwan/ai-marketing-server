package databases

import "github.com/jmoiron/sqlx"

type Database interface {
	GetConnection() *sqlx.DB
}
