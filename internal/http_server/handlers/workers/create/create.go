package create

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct { // @TODO : добавить валидацию
	Institute   string `json:"institute" validate:"required"`
	Surname     string `json:"surname" validate:"required"`
	Name        string `json:"name" validate:"required"`
	MiddleName  string `json:"middle_name,omitempty"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Cabinet     string `json:"cabinet" validate:"required"`
	Position    string `json:"position" validate:"required"`
	Department  string `json:"department" validate:"required"`
}

type Response struct {
	resp.Response
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
	) (int, error)
}

func New(ctx context.Context, log *slog.Logger, userCreater UserCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.create.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

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

		_, err = userCreater.CreateUser(
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
		)
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			msg := "user already exists"
			log.Info(msg, slog.String("email", req.Email))

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

		responseOk(w, r)
	}
}

func responseOk(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
