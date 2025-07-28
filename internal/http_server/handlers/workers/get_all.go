package workers

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/domain/models"
	"telephone-book/internal/lib/logger/sl"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AllUsersResponse struct {
	resp.Response
	Users []models.User `json:"users"`
}

type AllUsersGetter interface {
	GetAllUsers(ctx context.Context, institute string, department string) ([]models.User, error)
}

func GetAll(ctx context.Context, log *slog.Logger, allUsersGetter AllUsersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.read.GetAll"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}
		department := r.URL.Query().Get("department")
		if department == "" {
			msg := "department not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log = log.With(
			slog.String("institute", institute),
			slog.String("department", department),
		)

		users, err := allUsersGetter.GetAllUsers(ctx, institute, department)
		if err != nil {
			msg := "failed to get users"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("users retrieved successfully", slog.Int("count", len(users)))

		getResponseOk(w, r, users)
	}
}

func getResponseOk(w http.ResponseWriter, r *http.Request, users []models.User) {
	render.JSON(w, r, AllUsersResponse{
		Response: resp.OK(),
		Users:    users,
	})
}
