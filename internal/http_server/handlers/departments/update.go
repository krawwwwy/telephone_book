package departments

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UpdateRequest struct {
	Name     string   `json:"name"`
	Sections []string `json:"sections,omitempty"`
}

type UpdateResponse struct {
	resp.Response
}

type DepartmentUpdater interface {
	UpdateDepartment(
		ctx context.Context,
		institute string,
		oldName string,
		name string,
		sections []string,
	) error
}

func Update(ctx context.Context, log *slog.Logger, departmentUpdater DepartmentUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.departments.update.Update"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role != middleware.RoleAdmin {
			render.JSON(w, r, resp.Error("unauthorized: only authenticated admins can update workers"))
			return
		}
		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))

			return
		}
		oldName := r.URL.Query().Get("department")
		if oldName == "" {
			msg := "department name is not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))

			return
		}

		var req UpdateRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			msg := "failed to decode request body"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		err = departmentUpdater.UpdateDepartment(
			ctx,
			institute,
			oldName,
			req.Name,
			req.Sections,
		)

		if err != nil {
			msg := "failed to update department"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully updated",
			slog.String("old_name", oldName),
			slog.String("new_name", req.Name))

		responseOk(w, r)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, UpdateResponse{
		Response: resp.OK(),
	})
}
