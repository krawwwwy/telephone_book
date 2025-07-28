package workers

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/http_server/middleware"
	"telephone-book/internal/lib/logger/sl"
	resp "telephone-book/internal/lib/response"
	"telephone-book/internal/storage"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type DeleteResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
}

type UserDeleter interface {
	DeleteUser(
		ctx context.Context,
		institute string,
		email string,
	) error
}

// Delete удаляет работника
// @Summary Удалить работника
// @Tags workers
// @Produce json
// @Param institute query string true "Институт"
// @Param email query string true "Email работника"
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /workers [delete]
func Delete(ctx context.Context, log *slog.Logger, userDeleter UserDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.delete.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		email := r.URL.Query().Get("email")
		if email == "" {
			msg := "email not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		role := middleware.GetRole(r.Context(), log)
		if role != middleware.RoleAdmin {
			render.JSON(w, r, resp.Error("unauthorized: only admin can delete users"))
			return
		}

		err := userDeleter.DeleteUser(ctx, institute, email)
		if err != nil {
			if err == storage.ErrUserNotFound {
				msg := "user not found"
				log.Info(msg, slog.String("email", email))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			msg := "failed to delete user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully deleted",
			slog.String("email", email),
			slog.String("institute", institute))

		render.JSON(w, r, DeleteResponse{
			Status: resp.OK().Status,
			Error:  "",
		})
	}
}
