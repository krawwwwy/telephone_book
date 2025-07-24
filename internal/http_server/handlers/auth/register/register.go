package register

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/clients/sso/grpc"
	"telephone-book/internal/lib/logger/sl"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterResponse struct {
	resp.Response
	UserID int64 `json:"user_id"`
}

func New(ctx context.Context, ssoClient *grpc.Client, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role != middleware.RoleAdmin {
			log.Error("unauthorized: only admin can register users")
			render.JSON(w, r, resp.Error("unauthorized"))
			return
		}

		var req RegisterRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode register request", sl.Err(err))
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		userID, err := ssoClient.Register(r.Context(), req.Email, req.Password, req.Role)
		if err != nil {
			log.Error("failed to register user", sl.Err(err))
			render.JSON(w, r, resp.Error("registration failed"))
			return
		}

		log.Info("user registered successfully", slog.Int("user_id", int(userID)))
		responseOK(w, r, userID)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, ID int64) {
	render.JSON(w, r, RegisterResponse{
		Response: resp.OK(),
		UserID:   ID,
	})
}
