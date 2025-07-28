package departments

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/domain/models"
	"telephone-book/internal/lib/logger/sl"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type DepartmentsResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Список отделов
	Departments []models.Department `json:"departments"`
}

type SectionsResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// Список секций
	Sections []models.Section `json:"sections"`
}

type DepartmentsGetter interface {
	GetAllDepartments(ctx context.Context, institute string) ([]models.Department, error)
	GetSections(ctx context.Context, institute string, department string) ([]models.Section, error)
}

// GetAll возвращает список всех отделов
// @Summary Получить все отделы
// @Tags departments
// @Produce json
// @Param institute query string true "Институт"
// @Success 200 {object} DepartmentsResponse
// @Failure 400 {object} response.Response
// @Router /departments [get]
func GetAll(ctx context.Context, log *slog.Logger, departmnetsGetter DepartmentsGetter) http.HandlerFunc {
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

		departments, err := departmnetsGetter.GetAllDepartments(ctx, institute)
		if err != nil {
			msg := "failed to get departments"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("departments retrieved successfully", slog.Int("count", len(departments)))

		DepartmentsResponseOk(w, r, departments)
	}
}

// GetSections возвращает список секций отдела
// @Summary Получить секции отдела
// @Tags departments
// @Produce json
// @Param institute query string true "Институт"
// @Param department path string true "Название отдела"
// @Success 200 {object} SectionsResponse
// @Failure 400 {object} response.Response
// @Router /departments/{department} [get]
func GetSections(ctx context.Context, log *slog.Logger, departmnetsGetter DepartmentsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.departments.read.GetSections"

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

		department := chi.URLParam(r, "department")
		if department == "" {
			msg := "department not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log = log.With(
			slog.String("institute", institute),
			slog.String("department", department),
		)

		sections, err := departmnetsGetter.GetSections(ctx, institute, department)
		if err != nil {
			msg := "failed to get sections"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("sections retrieved successfully", slog.Int("count", len(sections)))

		SectionsResponseOk(w, r, sections)
	}
}
func DepartmentsResponseOk(w http.ResponseWriter, r *http.Request, departments []models.Department) {
	render.JSON(w, r, DepartmentsResponse{
		Status:      resp.OK().Status,
		Error:       "",
		Departments: departments,
	})
}

func SectionsResponseOk(w http.ResponseWriter, r *http.Request, sections []models.Section) {
	render.JSON(w, r, SectionsResponse{
		Status:   resp.OK().Status,
		Error:    "",
		Sections: sections,
	})
}
