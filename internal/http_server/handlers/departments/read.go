package departments

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AllDepartmentsResponse struct {
	resp.Response
	Departments []string `json:"departments"`
}

type AllDepartmentsGetter interface {
	GetAllDepartments(ctx context.Context, institute string) ([]string, error)
}

func GetAll(ctx context.Context, log *slog.Logger, allDepartmnetsGetter AllDepartmentsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.depaerments.read.GetAll"

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

		log = log.With(
			slog.String("institute", institute),
		)

		departments, err := allDepartmnetsGetter.GetAllDepartments(ctx, institute)
		if err != nil {
			msg := "failed to get departments"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("departments retrieved successfully", slog.Int("count", len(departments)))

		readResponseOk(w, r, departments)
	}
}

func readResponseOk(w http.ResponseWriter, r *http.Request, departments []string) {
	render.JSON(w, r, AllDepartmentsResponse{
		Response:    resp.OK(),
		Departments: departments,
	})
}
