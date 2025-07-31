package workers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"
	"time"

	middleware "telephone-book/internal/http_server/middleware"
	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const maxPhotoSizeWithPhoto = 5 * 1024 * 1024 // 5 MB

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// CreateWithPhoto создает нового работника с фотографией
// @Summary Создать работника с фотографией
// @Tags workers
// @Accept multipart/form-data
// @Produce json
// @Param institute formData string true "Институт"
// @Param surname formData string true "Фамилия"
// @Param name formData string true "Имя"
// @Param middle_name formData string false "Отчество"
// @Param email formData string true "Email"
// @Param phone_number formData string true "Рабочий телефон"
// @Param cabinet formData string false "Кабинет"
// @Param position formData string false "Должность"
// @Param department formData string false "Отдел"
// @Param section formData string false "Подотдел"
// @Param birth_date formData string false "Дата рождения (YYYY-MM-DD)"
// @Param description formData string false "Описание"
// @Param photo formData file false "Фотография (max 5MB)"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /workers/with-photo [post]
func CreateWithPhoto(ctx context.Context, log *slog.Logger, userCreater UserCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.create_with_photo.CreateWithPhoto"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

		role := middleware.GetRole(r.Context(), log)
		if role == middleware.RoleGuest {
			render.JSON(w, r, resp.Error("unauthorized: only authenticated users can create workers"))
			return
		}

		// Ограничиваем размер запроса
		r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSizeWithPhoto+1024*1024) // +1MB для остальных данных

		if err := r.ParseMultipartForm(maxPhotoSizeWithPhoto); err != nil {
			msg := "failed to parse multipart form"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Получаем обязательные поля
		institute := strings.TrimSpace(r.FormValue("institute"))
		surname := strings.TrimSpace(r.FormValue("surname"))
		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		phoneNumber := strings.TrimSpace(r.FormValue("phone_number"))

		if institute == "" {
			render.JSON(w, r, resp.Error("institute is required"))
			return
		}
		if surname == "" {
			render.JSON(w, r, resp.Error("surname is required"))
			return
		}
		if name == "" {
			render.JSON(w, r, resp.Error("name is required"))
			return
		}
		if email == "" {
			render.JSON(w, r, resp.Error("email is required"))
			return
		}
		if phoneNumber == "" {
			render.JSON(w, r, resp.Error("phone_number is required"))
			return
		}

		// Получаем необязательные поля
		middleName := strings.TrimSpace(r.FormValue("middle_name"))
		cabinet := strings.TrimSpace(r.FormValue("cabinet"))
		position := strings.TrimSpace(r.FormValue("position"))
		department := strings.TrimSpace(r.FormValue("department"))
		section := strings.TrimSpace(r.FormValue("section"))
		description := strings.TrimSpace(r.FormValue("description"))

		// Парсим дату рождения
		var birthDate time.Time
		birthDateStr := strings.TrimSpace(r.FormValue("birth_date"))
		if birthDateStr != "" {
			var err error
			birthDate, err = time.Parse("2006-01-02", birthDateStr)
			if err != nil {
				msg := "invalid birth_date format (use YYYY-MM-DD)"
				log.Error(msg, sl.Err(err))
				render.JSON(w, r, resp.Error(msg))
				return
			}
		}

		// Обрабатываем фотографию
		var photo []byte
		file, header, err := r.FormFile("photo")
		if err != nil && err != http.ErrMissingFile {
			msg := "failed to get photo file"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if file != nil {
			defer file.Close()

			// Проверяем размер файла
			if header.Size > maxPhotoSizeWithPhoto {
				msg := fmt.Sprintf("photo size too large (max %d MB)", maxPhotoSizeWithPhoto/(1024*1024))
				log.Warn(msg, slog.Int64("size", header.Size))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			// Проверяем тип файла по Content-Type
			contentType := header.Header.Get("Content-Type")
			if !allowedImageTypes[contentType] {
				// Дополнительная проверка по расширению файла
				ext := strings.ToLower(filepath.Ext(header.Filename))
				validExt := false
				switch ext {
				case ".jpg", ".jpeg", ".png", ".gif", ".webp":
					validExt = true
				}

				if !validExt {
					msg := "invalid file type (allowed: jpeg, jpg, png, gif, webp)"
					log.Warn(msg, slog.String("content_type", contentType), slog.String("filename", header.Filename))
					render.JSON(w, r, resp.Error(msg))
					return
				}
			}

			// Читаем файл
			photo, err = io.ReadAll(file)
			if err != nil {
				msg := "failed to read photo file"
				log.Error(msg, sl.Err(err))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			log.Info("photo received",
				slog.String("filename", header.Filename),
				slog.Int64("size", header.Size),
				slog.String("content_type", contentType),
			)
		}

		userID, err := userCreater.CreateUser(
			ctx,
			institute,
			surname,
			name,
			middleName,
			email,
			phoneNumber,
			cabinet,
			position,
			department,
			section,
			birthDate,
			description,
			photo,
		)

		if errors.Is(err, storage.ErrUserAlreadyExists) {
			msg := "user already exists"
			log.Warn(msg, slog.String("email", email))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if err != nil {
			msg := "failed to save user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("user successfully saved", slog.String("email", email), slog.Int("user_id", userID))

		createResponseOk(w, r, userID)
	}
}
