package manager

import (
	"github.com/jmoiron/sqlx"

	"github.com/B87/file-bridge/pkg/uri"
)

type URIHashStorage interface {
	// Get retrieves the URI for the given hash
	Get(uri uri.URI) (string, error)
	// Put stores the URI in the repository
	Put(uri uri.URI, srcHash string) error
}

type PGURIHashRepository struct {
	db *sqlx.DB
}

func NewPGURIHashRepository(db *sqlx.DB) *PGURIHashRepository {
	return &PGURIHashRepository{db: db}
}

func (r *PGURIHashRepository) Get(uri uri.URI) (string, error) {
	var hash string
	err := r.db.Get(&hash, "SELECT hash FROM uri_hash WHERE uri = $1", uri.String())
	return hash, err
}

func (r *PGURIHashRepository) Put(uri uri.URI, srcHash string) error {
	_, err := r.db.Exec("INSERT INTO uri_hash (uri, hash) VALUES ($1, $2)", uri.String(), uri.Name)
	return err
}
