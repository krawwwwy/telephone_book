package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/lib/logger/sl"
)

type Role int

const (
	RoleGuest Role = iota
	RoleUser
	RoleAdmin
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

func AuthMiddleware(ssoClient *grpc.Client, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log = log.With(
				slog.String("operation", "middleware.AuthMiddleware"),
				slog.String("request_id", r.Header.Get("X-Request-ID")),
			)

			token := extractToken(r, log)
			if token == "" {
				log.Debug("no token found in request")
				ctx := context.WithValue(r.Context(), roleKey, RoleGuest)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// TODO: получить секрет приложения (app secret) для валидации токена
			appSecret := "test-secret" // временно, нужно получать из конфига/БД
			userID, err := grpc.ParseUserIDFromToken(token, appSecret)
			if err != nil {
				log.Debug("failed to parse user ID from token", sl.Err(err))
				ctx := context.WithValue(r.Context(), roleKey, RoleGuest)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			isAdmin, err := ssoClient.IsAdmin(r.Context(), userID)
			role := RoleUser
			if err == nil && isAdmin {
				role = RoleAdmin
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request, log *slog.Logger) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Debug("no Authorization header found")
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		log.Debug("invalid Authorization header format", slog.String("header", authHeader))
		return ""
	}
	return parts[1]
}

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func GetRole(ctx context.Context, log *slog.Logger) Role {
	role, ok := ctx.Value(roleKey).(Role)
	if !ok {
		log.Debug("role not found in context, returning RoleGuest")
		return RoleGuest
	}
	return role
}
