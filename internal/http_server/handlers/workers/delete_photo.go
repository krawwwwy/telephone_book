package workers

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type DeletePhotoResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type PhotoDeleter interface {
	DeleteUserPhoto(ctx context.Context, institute string, email string) error
}

// DeletePhoto удаляет фотографию работника
// @Summary Удалить фотографию работника
// @Tags workers
// @Produce json
// @Param email path string true "Email работника"
// @Param institute query string true "Институт"
// @Success 200 {object} DeletePhotoResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /workers/{email}/photo [delete]
func DeletePhoto(ctx context.Context, log *slog.Logger, photoDeleter PhotoDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.delete_photo.DeletePhoto"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		// Проверяем роль пользователя - только админы могут удалять фото
		role := middleware.GetRole(r.Context(), log)
		if role != middleware.RoleAdmin {
			msg := "forbidden: only administrators can delete worker photos"
			log.Warn(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем email из URL path, обходя проблему chi с точками в email
		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) < 4 { // /workers/{email}/photo
			msg := "email not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Декодируем URL-encoded email
		encodedEmail := parts[2] // /workers/{email}/photo
		email, err := url.QueryUnescape(encodedEmail)
		if err != nil {
			msg := "invalid email format"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем institute из query параметров
		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("processing photo deletion request",
			slog.String("email", email),
			slog.String("institute", institute))

		// Удаляем фотографию из базы данных
		err = photoDeleter.DeleteUserPhoto(ctx, institute, email)
		if err != nil {
			if err == storage.ErrUserNotFound {
				msg := "user not found"
				log.Warn(msg, slog.String("email", email))
				render.JSON(w, r, resp.Error(msg))
				return
			}
			msg := "failed to delete user photo"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("photo deleted successfully",
			slog.String("email", email),
			slog.String("institute", institute))

		render.JSON(w, r, DeletePhotoResponse{
			Status: resp.OK().Status,
		})
	}
}
