package delete

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"
	resp "telephone-book/internal/lib/response"
	"telephone-book/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

type UserDeleter interface {
	DeleteUser(
		ctx context.Context,
		institute string,
		email string,
	) error
}

func New(ctx context.Context, log *slog.Logger, userDeleter UserDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.delete.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
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

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
