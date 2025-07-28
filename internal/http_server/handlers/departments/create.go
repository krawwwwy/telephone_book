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
	Name      string   `json:"name"`
	Institute string   `json:"institute"`
	Sections  []string `json:"sections,omitempty"` // Optional, can be used to specify sections within the department
}

type CreateResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// ID созданного отдела
	DepatmentID int `json:"user_id,omitempty"`
}

type DepatmentCreater interface {
	CreateDepartment(
		ctx context.Context,
		institute string,
		name string,
		sections []string, // Optional, can be used to specify sections within the department
	) (int, error)
}

// Create создает новый отдел
// @Summary Создать отдел
// @Tags departments
// @Accept json
// @Produce json
// @Param department body CreateRequest true "Данные отдела"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /departments [post]
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

		log = log.With(
			slog.String("institute", req.Institute),
			slog.String("department", req.Name),
		)

		log.Info("request body decoded", slog.Any("request", req))

		departmentID, err := departmentCreater.CreateDepartment(ctx, req.Institute, req.Name, req.Sections)
		if err != nil {
			log.Error("failed to create department", sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return

		}

		log.Info("department successfully saved")

		createResponseOk(w, r, departmentID)
	}
}

func createResponseOk(w http.ResponseWriter, r *http.Request, departmentID int) {
	render.JSON(w, r, CreateResponse{
		Status:      resp.OK().Status,
		Error:       "",
		DepatmentID: departmentID,
	})
}
