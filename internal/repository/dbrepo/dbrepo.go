package dbrepo

import (
	"database/sql"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo creates a new Postgres repository. Returns a DatabaseRepo
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
