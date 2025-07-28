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
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Список работников
	Users []models.User `json:"users"`
}

type AllUsersGetter interface {
	GetAllUsers(ctx context.Context, institute string, department string, section string) ([]models.User, error)
}

// GetAll возвращает всех работников отдела или секции
// @Summary Получить всех работников
// @Tags workers
// @Produce json
// @Param institute query string true "Институт"
// @Param department query string false "Отдел"
// @Param section query string false "Секция"
// @Success 200 {object} AllUsersResponse
// @Failure 400 {object} response.Response
// @Router /workers/all [post]
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

		// department может быть пустым - тогда вернем всех сотрудников института
		department := r.URL.Query().Get("department")
		// section может быть пустым - тогда вернем всех сотрудников отдела
		section := r.URL.Query().Get("section")

		log = log.With(
			slog.String("institute", institute),
			slog.String("department", department),
			slog.String("section", section),
		)

		users, err := allUsersGetter.GetAllUsers(ctx, institute, department, section)
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
		Status: resp.OK().Status,
		Error:  "",
		Users:  users,
	})
}
