package filters

import (
	"context"
	"net/http"

	"github.com/otiai10/marmoset"
	"gopkg.in/mgo.v2"
)

// MongoFilter handle MongoDB connection, to prevent "too many open files"
type MongoFilter struct {
	marmoset.Filter
	root *mgo.Session
}

// InitMongoFilter ...
func InitMongoFilter(root *mgo.Session) *MongoFilter {
	return &MongoFilter{root: root}
}

func (f *MongoFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := f.root.Clone()
	ctx := marmoset.Context().Get(r)
	marmoset.Context().Set(r, context.WithValue(ctx, mongoSessionKey, session))
	defer session.Close()
	f.Next.ServeHTTP(w, r)
}
