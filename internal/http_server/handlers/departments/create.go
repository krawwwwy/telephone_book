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

type CreateRequest struct {
	Name string `json:"name"`
}

type CreateResponse struct {
	resp.Response
	DepatmentID int `json:"user_id,omitempty"`
}

type DepatmentCreater interface {
	CreateDepartment(
		ctx context.Context,
		institute string,
		name string,
	) (int, error)
}

func Create(ctx context.Context, log *slog.Logger, departmentCreater DepatmentCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.departments.create.Create"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role != middleware.RoleAdmin {
			render.JSON(w, r, resp.Error("unauthorized: only admins can create departments"))
			return
		}

		var req CreateRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			msg := "failed to decode request body"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute parameter is required"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		departmentID, err := departmentCreater.CreateDepartment(ctx, institute, req.Name)
		if err != nil {
			log.Error("failed to create department", slog.String("institute", institute), sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return

		}

		log.Info("department successfully saved", slog.String("department", req.Name))

		createResponseOk(w, r, departmentID)
	}
}

func createResponseOk(w http.ResponseWriter, r *http.Request, departmentID int) {
	render.JSON(w, r, CreateResponse{
		Response:    resp.OK(),
		DepatmentID: departmentID,
	})
}
