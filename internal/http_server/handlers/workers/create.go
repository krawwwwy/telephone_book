package workers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"
	"time"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

const maxPhotoSizeCreate = 5 * 1024 * 1024 // 5 MB

type CreateRequest struct {
	Institute   string    `json:"institute" validate:"required"`
	Surname     string    `json:"surname" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	MiddleName  string    `json:"middle_name,omitempty"`
	Email       string    `json:"email" validate:"required,email"`
	PhoneNumber string    `json:"phone_number" validate:"required"`
	Cabinet     string    `json:"cabinet,omitempty"`
	Position    string    `json:"position,omitempty"`
	Department  string    `json:"department,omitempty"`
	Section     string    `json:"section,omitempty"`
	BirthDate   time.Time `json:"birth_date,omitempty"`
	Description string    `json:"description,omitempty"`
}

type CreateResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// ID созданного работника
	UserID int `json:"user_id,omitempty"`
}

type UserCreater interface {
	CreateUser(
		ctx context.Context,
		institute string,
		surname string,
		name string,
		middlename string,
		email string,
		phoneNumber string,
		cabinet string,
		position string,
		department string,
		section string,
		birthDate time.Time,
		description string,
		photo []byte,
	) (int, error)
}

// Create создает нового работника
// @Summary Создать работника
// @Tags workers
// @Accept json
// @Produce json
// @Param worker body CreateRequest true "Данные работника"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /workers [post]
func Create(ctx context.Context, log *slog.Logger, userCreater UserCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.create.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role == middleware.RoleGuest {
			render.JSON(w, r, resp.Error("unauthorized: only authenticated users can create workers"))
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

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			msg := "invalid request"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		userID, err := userCreater.CreateUser(
			ctx,
			req.Institute,
			req.Surname,
			req.Name,
			req.MiddleName,
			req.Email,
			req.PhoneNumber,
			req.Cabinet,
			req.Position,
			req.Department,
			req.Section,
			req.BirthDate,
			req.Description,
			nil, // Без фотографии
		)

		if errors.Is(err, storage.ErrUserAlreadyExists) {
			msg := "user already exists"
			log.Warn(msg, slog.String("email", req.Email))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if err != nil {
			msg := "failed to save user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully saved", slog.String("email", req.Email))

		createResponseOk(w, r, userID)
	}
}

func createResponseOk(w http.ResponseWriter, r *http.Request, userID int) {
	render.JSON(w, r, CreateResponse{
		Status: resp.OK().Status,
		Error:  "",
		UserID: userID,
	})
}
