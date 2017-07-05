package app

import (
	"net/http"
	"os"

	"gopkg.in/mgo.v2"

	"github.com/otiai10/fastpot-api/controllers/v0"
	"github.com/otiai10/fastpot-api/filters"
	"github.com/otiai10/marmoset"
)

func init() {

	session, err := mgo.Dial(os.Getenv("MONGODB_URI"))
	if err != nil {
		panic(err)
	}
	// defer session.Close()

	mf := filters.InitMongoFilter(session)
	lf := filters.InitLogFilter()
	cf := new(marmoset.ContextFilter)

	unauthorized := marmoset.NewRouter()
	unauthorized.GET("/0/status", v0.Status)

	authorized := marmoset.NewRouter()
	authorized.Apply(filters.InitializeAuthFilter())

	root := marmoset.NewRouter()
	root.Apply(lf, mf, cf)
	root.Subrouter(unauthorized)
	root.Subrouter(authorized)
	http.Handle("/", root)
}
