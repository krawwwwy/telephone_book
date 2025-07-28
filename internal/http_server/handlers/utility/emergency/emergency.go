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
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Список срочных служб
	Services []models.Service `json:"services"`
}

// New возвращает список срочных служб
// @Summary Получить срочные службы
// @Tags emergency
// @Produce json
// @Success 200 {object} EmergencyResponse
// @Failure 400 {object} response.Response
// @Router /emergency [get]
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
		Status:   resp.OK().Status,
		Error:    "",
		Services: services,
	})
}
