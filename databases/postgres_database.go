package databases

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ai-marketing/ai-marketing-server/config"
)

type postgresDatabase struct {
	db *sqlx.DB
}

var (
	postgresDatabaseInstance *postgresDatabase
	once                     sync.Once
)

func NewPostgresDatabase(conf *config.Database) Database {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
			conf.Host,
			conf.Port,
			conf.User,
			conf.Password,
			conf.DBName,
			conf.SSLMode,
			conf.Schema,
		)

		conn, err := sqlx.Open(conf.Driver, dsn)
		if err != nil {
			panic(err)
		}

		conn.SetConnMaxLifetime(3 * time.Minute)
		conn.SetMaxOpenConns(10)
		conn.SetMaxIdleConns(10)

		log.Printf("Connected to database %s", conf.DBName)

		postgresDatabaseInstance = &postgresDatabase{db: conn}
	})

	return postgresDatabaseInstance
}

func (db *postgresDatabase) GetConnection() *sqlx.DB {
	return db.db
}
