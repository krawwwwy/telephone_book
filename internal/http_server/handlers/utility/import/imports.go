package imports

import (
	"context"
	"log/slog"
	"mime/multipart"
	"net/http"
	"telephone-book/internal/domain/models"
	middleware "telephone-book/internal/http_server/middleware"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/lib/parser"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UserImporter interface {
	ImportUsers(ctx context.Context, isntitue string, users []models.User) error
}

func New(ctx context.Context, log *slog.Logger, userCreater UserImporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.utility.imports.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role == middleware.RoleGuest {
			render.JSON(w, r, resp.Error("unauthorized: only authenticated users can import data"))
			return
		}

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute parameter is required"
			log.Error(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		err := r.ParseMultipartForm(100 << 20) // 100 MB limit
		if err != nil {
			log.Error("failed to parse multipart form", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to parse form data"))

			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			log.Error("failed to get file from form", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to get file from form"))

			return
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Warn("failed to close file", sl.Err(err))
			}
		}(file)

		users, err := parser.Excel(file)
		if err != nil {
			log.Error("failed to parse excel file", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to parse excel file"))

			return
		}

		err = userCreater.ImportUsers(ctx, institute, users)
		if err != nil {
			msg := "failed to import users"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		render.JSON(w, r, resp.OK())
	}
}
