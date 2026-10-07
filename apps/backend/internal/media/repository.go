package media

import (
	"github.com/playingwithclouds/veil/internal/db"
)

type Repository struct {
	database *db.DB
}

func NewRepository(database *db.DB) *Repository {
	return &Repository{database: database}
}
