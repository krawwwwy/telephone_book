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
	Search(ctx context.Context, institute string, department string, section string, query string) ([]models.User, error)
}

type AllUsersResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Список пользователей
	Users []models.User `json:"users"`
}

// New ищет пользователей по параметрам
// @Summary Поиск пользователей
// @Tags search
// @Produce json
// @Param institute query string true "Институт"
// @Param department query string false "Отдел"
// @Param section query string false "Секция"
// @Param query query string true "Строка поиска"
// @Success 200 {object} AllUsersResponse
// @Failure 400 {object} response.Response
// @Router /search [get]
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

		department := r.URL.Query().Get("department")
		section := r.URL.Query().Get("section") // может быть пустым (тогда ищем по всему институту)

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

		users, err := usersSearcher.Search(ctx, institute, department, section, query)
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
		Status: resp.OK().Status,
		Error:  "",
		Users:  users,
	})
}
