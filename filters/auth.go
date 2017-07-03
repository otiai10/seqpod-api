package filters

import (
	"net/http"

	"github.com/otiai10/marmoset"
)

// AuthFilter ...
type AuthFilter struct {
	marmoset.Filter
}

// AuthCtxKey ...
type AuthCtxKey string

// InitializeAuthFilter ...
func InitializeAuthFilter() *AuthFilter {
	return &AuthFilter{}
}

// ServeHTTP ...
func (f *AuthFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// c, err := r.Cookie("chant_identity_token")
	// if err != nil {
	// 	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	// 	return
	// }
	// user, err := models.DecodeUser(c.Value, os.Getenv("JWT_SALT"))
	// if err != nil {
	// 	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	// 	return
	// }
	// if user == nil {
	// 	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	// 	return
	// }
	// marmoset.Context().Set(r, context.WithValue(ctx, AuthKey, user))
	// f.Next.ServeHTTP(w, r)
	f.Next.ServeHTTP(w, r)
	return

}

// RequestUser returns context user which auth filter detected and has set
// func RequestUser(r *http.Request) *models.User {
// 	user, ok := marmoset.Context().Get(r).Value(AuthKey).(*models.User)
// 	if !ok {
// 		return nil
// 	}
// 	return user
// }
