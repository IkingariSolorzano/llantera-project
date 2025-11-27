package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/llantera/hex/internal/config"
)

// Open crea una conexión *sql.DB siguiendo los parámetros definidos en config.Config.
func Open(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	if cfg.DBConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(time.Hour)
	}

	return db, db.Ping()
}
