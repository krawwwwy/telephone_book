package search

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/domain/models"

	"telephone-book/internal/lib/logger/sl"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UsersSearcher interface {
	Search(ctx context.Context, institute string, department string, query string) ([]models.User, error)
}

type AllUsersResponse struct {
	resp.Response
	Users []models.User `json:"users"`
}

func New(ctx context.Context, log *slog.Logger, usersSearcher UsersSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.search.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		department := r.URL.Query().Get("department") // может быть пустым (тогда ищем по всему институту)
		query := r.URL.Query().Get("query")
		if query == "" {
			msg := "query not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log = log.With(
			slog.String("institute", institute),
			slog.String("department", department),
			slog.String("query", query),
		)

		users, err := usersSearcher.Search(ctx, institute, department, query)
		if err != nil {
			msg := "failed to search users"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("users found", slog.Int("count", len(users)))

		responseOk(w, r, users)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request, users []models.User) {
	render.JSON(w, r, AllUsersResponse{
		Response: resp.OK(),
		Users:    users,
	})
}
