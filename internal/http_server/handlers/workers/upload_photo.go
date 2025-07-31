package workers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const maxPhotoSizeUpload = 5 * 1024 * 1024 // 5 MB

var allowedImageTypesUpload = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type UploadPhotoResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type PhotoUploader interface {
	UpdateUserPhoto(ctx context.Context, institute string, email string, photo []byte) error
	GetUserPhoto(ctx context.Context, institute string, email string) ([]byte, error)
}

// UploadPhoto загружает фотографию работника (только если у него нет фото)
// @Summary Загрузить фотографию работника
// @Tags workers
// @Accept multipart/form-data
// @Produce json
// @Param email path string true "Email работника"
// @Param institute query string true "Институт"
// @Param photo formData file true "Фотография"
// @Success 200 {object} UploadPhotoResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /workers/{email}/photo [post]
func UploadPhoto(ctx context.Context, log *slog.Logger, photoUploader PhotoUploader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.upload_photo.UploadPhoto"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		// Проверяем роль пользователя - авторизованные пользователи могут загружать фото
		role := middleware.GetRole(r.Context(), log)
		if role == middleware.RoleGuest {
			msg := "unauthorized: only authenticated users can upload worker photos"
			log.Warn(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем email из URL path, обходя проблему chi с точками в email
		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) < 4 { // /workers/{email}/photo
			msg := "email not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Декодируем URL-encoded email
		encodedEmail := parts[2] // /workers/{email}/photo
		email, err := url.QueryUnescape(encodedEmail)
		if err != nil {
			msg := "invalid email format"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем institute из query параметров
		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("processing photo upload request",
			slog.String("email", email),
			slog.String("institute", institute))

		// Проверяем, что у пользователя нет фото
		existingPhoto, err := photoUploader.GetUserPhoto(ctx, institute, email)
		if err != nil && err != storage.ErrUserNotFound {
			msg := "failed to check existing photo"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if len(existingPhoto) > 0 {
			msg := "user already has a photo, use PUT method to update"
			log.Warn(msg, slog.String("email", email))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Ограничиваем размер запроса
		r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSizeUpload+1024*1024)

		// Парсим multipart form
		if err := r.ParseMultipartForm(maxPhotoSizeUpload); err != nil {
			msg := "failed to parse multipart form"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем файл фотографии
		file, header, err := r.FormFile("photo")
		if err != nil {
			msg := "photo file is required"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}
		defer file.Close()

		// Проверяем размер файла
		if header.Size > maxPhotoSizeUpload {
			msg := "photo file is too large (max 5MB)"
			log.Warn(msg, slog.Int64("size", header.Size))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Проверяем тип файла
		contentType := header.Header.Get("Content-Type")
		if !allowedImageTypesUpload[contentType] {
			msg := "unsupported image type"
			log.Warn(msg, slog.String("content_type", contentType))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Читаем содержимое файла
		photo, err := io.ReadAll(file)
		if err != nil {
			msg := "failed to read photo file"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Загружаем фотографию в базу данных
		if err = photoUploader.UpdateUserPhoto(ctx, institute, email, photo); err != nil {
			if err == storage.ErrUserNotFound {
				msg := "user not found"
				log.Warn(msg, slog.String("email", email))
				render.JSON(w, r, resp.Error(msg))
				return
			}
			msg := "failed to upload user photo"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("photo uploaded successfully",
			slog.String("email", email),
			slog.String("institute", institute),
			slog.Int("photo_size", len(photo)))

		render.JSON(w, r, UploadPhotoResponse{
			Status: resp.OK().Status,
		})
	}
}
