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
	resp.Response
}

type DepartmentDeleter interface {
	DeleteDepartment(
		ctx context.Context,
		institute string,
		name string,
	) error
}

func Delete(ctx context.Context, log *slog.Logger, departmentDeleter DepartmentDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.deparments.delete.Delete"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		name := r.URL.Query().Get("name")
		if name == "" {
			msg := "name not specified"
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

		err := departmentDeleter.DeleteDepartment(ctx, institute, name)
		if err != nil {
			msg := "failed to delete user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully deleted",
			slog.String("department", name),
			slog.String("institute", institute))

		render.JSON(w, r, DeleteResponse{
			Response: resp.OK(),
		})
	}
}
