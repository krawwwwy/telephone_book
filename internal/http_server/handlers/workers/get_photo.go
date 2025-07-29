package workers

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"telephone-book/internal/lib/logger/sl"
	"telephone-book/internal/storage"
	"time"

	resp "telephone-book/internal/lib/response"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// GetPhoto возвращает фотографию пользователя
// @Summary Получить фотографию пользователя
// @Tags workers
// @Produce image/jpeg,image/png,image/gif,image/webp
// @Param email path string true "Email пользователя"
// @Param institute query string true "Институт"
// @Success 200 {file} binary "Фотография пользователя"
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /workers/{email}/photo [get]
func GetPhoto(ctx context.Context, log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workers.get_photo.GetPhoto"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", chimw.GetReqID(r.Context())),
		)

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

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute not specified"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		user, err := userGetter.GetUserPhoto(ctx, institute, email)
		if err != nil {
			if err == storage.ErrUserNotFound {
				msg := "user not found"
				log.Info(msg, slog.String("email", email))
				render.JSON(w, r, resp.Error(msg))
				return
			}

			msg := "failed to get user"
			log.Error(msg, sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		if len(user) == 0 {
			msg := "user has no photo"
			log.Info(msg, slog.String("email", email))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		// Определяем тип контента по первым байтам изображения
		contentType := detectImageType(user)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Генерируем ETag для кеширования
		hash := md5.Sum(user)
		etag := fmt.Sprintf(`"%x"`, hash)

		// Проверяем If-None-Match для кеширования
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(user)))
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=3600") // Кеш на 1 час
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))

		// Отправляем изображение
		w.WriteHeader(http.StatusOK)
		w.Write(user)

		log.Info("photo served successfully",
			slog.String("email", email),
			slog.Int("size", len(user)),
			slog.String("content_type", contentType),
		)
	}
}

// detectImageType определяет тип изображения по магическим байтам
func detectImageType(data []byte) string {
	if len(data) < 12 {
		return ""
	}

	// JPEG
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG
	if len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png"
	}

	// GIF
	if len(data) >= 6 &&
		((data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 && data[4] == 0x37 && data[5] == 0x61) ||
			(data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 && data[4] == 0x39 && data[5] == 0x61)) {
		return "image/gif"
	}

	// WebP
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "image/webp"
	}

	return ""
}
