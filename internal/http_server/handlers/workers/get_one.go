package workers

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/domain/models"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type GetResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Работник
	User models.User `json:"user"`
}

type UserGetter interface {
	GetUserByEmail(ctx context.Context, institute string, email string) (models.User, error)
}

// GetByEmail возвращает работника по email
// @Summary Получить работника по email
// @Tags workers
// @Produce json
// @Param email path string true "Email работника"
// @Param institute query string true "Институт"
// @Success 200 {object} GetResponse
// @Failure 400 {object} response.Response
// @Router /workers/{email} [get]
func GetByEmail(ctx context.Context, log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.read.GetByEmail"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		email := chi.URLParam(r, "email")
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

		user, err := userGetter.GetUserByEmail(ctx, institute, email)
		if err != nil {
			if err == storage.ErrUserNotFound {
				msg := "user not found"
				log.Info(msg, slog.String("email", email))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			msg := "failed to get user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user retrieved successfully", slog.String("email", email))

		render.JSON(w, r, GetResponse{
			Status: resp.OK().Status,
			Error:  "",
			User:   user,
		})
	}
}
