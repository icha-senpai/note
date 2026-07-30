package mongomgo

import (
	"github.com/icha-senpai/note/third_party/forks/github/gin-contrib/sessions"
	"github.com/icha-senpai/note/third_party/forks/github/globalsign/mgo"
	"github.com/icha-senpai/note/third_party/forks/github/kidstuff/mongostore"
)

var _ sessions.Store = (*store)(nil)

func NewStore(c *mgo.Collection, maxAge int, ensureTTL bool, keyPairs ...[]byte) sessions.Store {
	return &store{mongostore.NewMongoStore(c, maxAge, ensureTTL, keyPairs...)}
}

type store struct {
	*mongostore.MongoStore
}

func (c *store) Options(options sessions.Options) {
	c.MongoStore.Options = options.ToGorillaOptions()
}
