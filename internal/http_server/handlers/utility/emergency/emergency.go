package emergency

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

type EmergencyProvider interface {
	Emergency(ctx context.Context) ([]models.Service, error)
}

type EmergencyResponse struct {
	resp.Response
	Services []models.Service `json:"services"`
}

func New(ctx context.Context, log *slog.Logger, emergencyProvider EmergencyProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.utility.emergency.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		services, err := emergencyProvider.Emergency(ctx)
		if err != nil {
			log.Error("failed to get emergency services", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to retrieve emergency services"))
			return
		}

		log.Info("emergency services retrieved successfully", slog.Int("count", len(services)))

		responseOk(w, r, services)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request, services []models.Service) {
	render.JSON(w, r, EmergencyResponse{
		Response: resp.OK(),
		Services: services,
	})
}
