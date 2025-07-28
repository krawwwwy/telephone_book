package workers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"telephone-book/internal/domain/models"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"
	"time"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

const maxPhotoSize = 5 * 1024 * 1024 // 5 MB

type UpdateRequest struct {
	Institute   string    `json:"institute" validate:"required"`
	Surname     string    `json:"surname" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	MiddleName  string    `json:"middle_name,omitempty"`
	Email       string    `json:"email" validate:"required,email"`
	PhoneNumber string    `json:"phone_number" validate:"required"`
	Cabinet     string    `json:"cabinet" validate:"required"`
	Position    string    `json:"position" validate:"required"`
	Department  string    `json:"department" validate:"required"`
	Section     string    `json:"section,omitempty"`
	BirthDate   time.Time `json:"birth_date,omitempty"`
	Description string    `json:"description,omitempty"`
	Photo       []byte    `json:"photo,omitempty"`
}

type UpdateResponse struct {
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
}

type UserUpdater interface {
	UpdateUser(
		ctx context.Context,
		institute string,
		oldEmail string,
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
	) error
	GetUserByEmail(ctx context.Context, institute string, email string) (models.User, error)
}

// Update обновляет работника
// @Summary Обновить работника
// @Tags workers
// @Accept json
// @Produce json
// @Param institute query string true "Институт"
// @Param email query string true "Email работника"
// @Param worker body UpdateRequest true "Новые данные работника"
// @Success 200 {object} UpdateResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /workers [put]
func Update(ctx context.Context, log *slog.Logger, userUpdater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.update.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role == middleware.RoleGuest {
			render.JSON(w, r, resp.Error("unauthorized: only authenticated users can update workers"))
			return
		}
		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))

			return
		}
		oldEmail := r.URL.Query().Get("email")
		if oldEmail == "" {
			msg := "email not specified"
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

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			msg := "invalid request"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Только admin может изменять уже заполненные поля, user может только пустые
		if role == middleware.RoleUser {
			// Получаем текущие данные пользователя
			user, errGet := userUpdater.GetUserByEmail(ctx, institute, oldEmail)
			if errGet != nil {
				msg := "failed to get current user data"
				log.Error(msg, sl.Err(errGet))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			// Проверяем, что пользователь меняет только пустые поля
			var forbiddenFields []string
			if user.Surname != "" && user.Surname != req.Surname {
				forbiddenFields = append(forbiddenFields, "surname")
			}
			if user.Name != "" && user.Name != req.Name {
				forbiddenFields = append(forbiddenFields, "name")
			}
			if user.MiddleName != "" && user.MiddleName != req.MiddleName {
				forbiddenFields = append(forbiddenFields, "middle_name")
			}
			if user.Email != "" && user.Email != req.Email {
				forbiddenFields = append(forbiddenFields, "email")
			}
			if user.PhoneNumber != "" && user.PhoneNumber != req.PhoneNumber {
				forbiddenFields = append(forbiddenFields, "phone_number")
			}
			if user.Cabinet != "" && user.Cabinet != req.Cabinet {
				forbiddenFields = append(forbiddenFields, "cabinet")
			}
			if user.Position != "" && user.Position != req.Position {
				forbiddenFields = append(forbiddenFields, "position")
			}
			if user.Department != "" && user.Department != req.Department {
				forbiddenFields = append(forbiddenFields, "department")
			}
			if user.Section != "" && user.Section != req.Section {
				forbiddenFields = append(forbiddenFields, "section")
			}
			if !user.BirthDate.IsZero() && !user.BirthDate.Equal(req.BirthDate) {
				forbiddenFields = append(forbiddenFields, "birth_date")
			}
			if user.Description != "" && user.Description != req.Description {
				forbiddenFields = append(forbiddenFields, "description")
			}

			if len(forbiddenFields) > 0 {
				msg := "user can only update empty fields: "
				log.Warn(msg)
				render.JSON(w, r, resp.Error(msg))
				return
			}
		}

		err = userUpdater.UpdateUser(
			ctx,
			req.Institute,
			oldEmail,
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

		if errors.Is(err, storage.ErrUserNotFound) {
			msg := "user not found"
			log.Warn(msg, slog.String("email", oldEmail))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if err != nil {
			msg := "failed to update user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully updated",
			slog.String("old_email", oldEmail),
			slog.String("new_email", req.Email))

		responseOk(w, r)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, UpdateResponse{
		Status: resp.OK().Status,
		Error:  "",
	})
}
