package repository

import "github.com/jmoiron/sqlx"

type IRepository interface {
}

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) IRepository {
	return Repository{
		DB: db,
	}
}
