package gorm

import (
	"time"

	"github.com/icha-senpai/note/third_party/forks/github/gin-contrib/sessions"
	"github.com/icha-senpai/note/third_party/forks/github/wader/gormstore/v2"
	"github.com/icha-senpai/note/third_party/forks/external/gorm.io/gorm"
)

type Store interface {
	sessions.Store
}

func NewStore(d *gorm.DB, expiredSessionCleanup bool, keyPairs ...[]byte) Store {
	s := gormstore.New(d, keyPairs...)
	if expiredSessionCleanup {
		quit := make(chan struct{})
		go s.PeriodicCleanup(1*time.Hour, quit)
	}
	return &store{s}
}

type store struct {
	*gormstore.Store
}

func (s *store) Options(options sessions.Options) {
	s.SessionOpts = options.ToGorillaOptions()
}
