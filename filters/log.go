package filters

import (
	"log"
	"net/http"

	"github.com/otiai10/marmoset"
)

// LogFilter ...
type LogFilter struct {
	marmoset.Filter
	Level int
}

// InitLogFilter ...
func InitLogFilter() *LogFilter {
	return &LogFilter{}
}

func (f *LogFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.String())
	f.Next.ServeHTTP(w, r)
}
