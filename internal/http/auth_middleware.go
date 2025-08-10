package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/Ali0NAL/talentpass/internal/auth"
)

type ctxKey string

const ctxUserIDKey ctxKey = "userID"

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxUserIDKey)
	id, ok := v.(int64)
	return id, ok
}

// RequireAuth: Geçerli Bearer JWT yoksa 401 döner.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		claims, err := auth.Parse(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		// userID'yi context'e ekle
		ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
