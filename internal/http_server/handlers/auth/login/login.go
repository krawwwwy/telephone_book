package login

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/lib/logger/sl"

	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse ответ на успешный вход
type LoginResponse struct {
	// Response базовая структура ответа
	// Статус ответа: Ok или Error. Пример: "Ok"
	Status string `json:"status"`
	// Сообщение об ошибке, если есть. Пример: "invalid request"
	Error string `json:"error,omitempty"`
	// JWT токен. Пример: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	Token string `json:"token"`
}

const AppID = 1

// New авторизует пользователя
// @Summary Вход пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Данные для входа"
// @Success 200 {object} login.LoginResponse
// @Failure 400 {object} response.Response
// @Router /auth/login [post]
func New(ctx context.Context, ssoClient *grpc.Client, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.login.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		var req LoginRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode login request", sl.Err(err))
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		token, err := ssoClient.Login(r.Context(), req.Email, req.Password, AppID)
		if err != nil {
			log.Error("failed to login", sl.Err(err))
			render.JSON(w, r, resp.Error("login failed"))
			return
		}
		log.Info("user logged in successfully", slog.String("email", req.Email))
		responseOK(w, r, token)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, token string) {
	render.JSON(w, r, LoginResponse{
		Status: resp.OK().Status,
		Error:  "",
		Token:  token,
	})
}
