package user_info

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/lib/logger/sl"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UserInfoResponse struct {
	Status string `json:"status"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Error  string `json:"error,omitempty"`
}

func UserInfo(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.user_info.UserInfo"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn("no authorization header")
			render.JSON(w, r, resp.Error("unauthorized"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Warn("invalid authorization header format")
			render.JSON(w, r, resp.Error("unauthorized"))
			return
		}

		token := parts[1]

		// Извлекаем email из токена (используем тот же секрет что и в middleware)
		email, err := grpc.ParseEmailFromToken(token, "test-secret")
		if err != nil {
			log.Error("failed to parse email from token", sl.Err(err))
			render.JSON(w, r, resp.Error("invalid token"))
			return
		}

		role := middleware.GetRole(r.Context(), log)

		var roleString string
		switch role {
		case middleware.RoleGuest:
			roleString = "guest"
		case middleware.RoleUser:
			roleString = "user"
		case middleware.RoleAdmin:
			roleString = "admin"
		default:
			roleString = "guest"
		}

		log.Info("user info retrieved",
			slog.String("email", email),
			slog.String("role", roleString))

		render.JSON(w, r, UserInfoResponse{
			Status: resp.OK().Status,
			Email:  email,
			Role:   roleString,
		})
	}
}
