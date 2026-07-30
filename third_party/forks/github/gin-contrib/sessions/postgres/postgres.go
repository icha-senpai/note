package postgres

import (
	"database/sql"

	"github.com/icha-senpai/note/third_party/forks/github/antonlindstrom/pgstore"
	"github.com/icha-senpai/note/third_party/forks/github/gin-contrib/sessions"
)

type Store interface {
	sessions.Store
}

type store struct {
	*pgstore.PGStore
}

var _ Store = new(store)

func NewStore(db *sql.DB, keyPairs ...[]byte) (Store, error) {
	p, err := pgstore.NewPGStoreFromPool(db, keyPairs...)
	if err != nil {
		return nil, err
	}

	return &store{p}, nil
}

func (s *store) Options(options sessions.Options) {
	s.PGStore.Options = options.ToGorillaOptions()
}
