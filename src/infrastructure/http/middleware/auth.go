package middleware

import (
	"context"
	"net/http"
	"strings"

	jwtutil "apiGolan/src/infrastructure/jwt"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

// Auth valida el token JWT del header Authorization: Bearer <token>
func Auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Permitir preflight CORS
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, `{"error":"token requerido"}`, http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwtutil.Validate(tokenStr)
        if err != nil {
            http.Error(w, `{"error":"token inválido o expirado"}`, http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// OnlyHost rechaza la petición si el usuario no es host
func OnlyHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserClaimsKey).(*jwtutil.Claims)
		if !ok || claims.Role != "host" {
			http.Error(w, `{"error":"solo el host puede realizar esta acción"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
