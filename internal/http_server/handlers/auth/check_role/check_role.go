package check_role

import (
	"context"
	"log/slog"
	"net/http"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type CheckRoleResponse struct {
	Status string `json:"status"`
	Role   string `json:"role"`
	Error  string `json:"error,omitempty"`
}

// CheckRole проверяет роль пользователя
// @Summary Проверить роль пользователя
// @Tags auth
// @Produce json
// @Success 200 {object} CheckRoleResponse
// @Failure 401 {object} response.Response
// @Router /auth/check-role [get]
func CheckRole(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.check_role.CheckRole"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

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

		log.Info("role checked", slog.String("role", roleString))

		render.JSON(w, r, CheckRoleResponse{
			Status: resp.OK().Status,
			Role:   roleString,
		})
	}
}
