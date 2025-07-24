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

type LoginResponse struct {
	resp.Response
	Token string `json:"token"`
}

const AppID = 1

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
		Response: resp.OK(),
		Token:    token,
	})
}
