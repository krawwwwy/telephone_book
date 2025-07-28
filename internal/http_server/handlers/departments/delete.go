package departments

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/http_server/middleware"
	"telephone-book/internal/lib/logger/sl"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type DeleteResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
}

type DepartmentDeleter interface {
	DeleteDepartment(
		ctx context.Context,
		institute string,
		name string,
	) error
}

// Delete удаляет отдел
// @Summary Удалить отдел
// @Tags departments
// @Produce json
// @Param institute query string true "Институт"
// @Param department query string true "Название отдела"
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /departments [delete]
func Delete(ctx context.Context, log *slog.Logger, departmentDeleter DepartmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.deparments.delete.Delete"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		department := r.URL.Query().Get("department")
		if department == "" {
			msg := "department not specified"
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
			render.JSON(w, r, resp.Error("unauthorized: only admin can delete departments"))
			return
		}

		err := departmentDeleter.DeleteDepartment(ctx, institute, department)
		if err != nil {
			msg := "failed to delete user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully deleted",
			slog.String("department", department),
			slog.String("institute", institute))

		render.JSON(w, r, DeleteResponse{
			Status: resp.OK().Status,
			Error:  "",
		})
	}
}
