package birthday

import (
	"context"
	"log/slog"
	"net/http"
	"telephone-book/internal/lib/logger/sl"

	resp "telephone-book/internal/lib/response"

	"github.com/go-chi/render"
)

// Tomorrow возвращает список пользователей, у которых день рождения завтра
// @Summary Дни рождения завтра
// @Tags birthday
// @Produce json
// @Param institute query string true "Институт"
// @Success 200 {array} models.User
// @Failure 400 {object} response.Response
// @Router /birthday/tomorrow [get]
func Tomorrow(ctx context.Context, log *slog.Logger, birthdayGetter BirthdayGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.utility.birthday.New"

		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", r.Header.Get("X-Request-ID")),
		)

		institute := r.URL.Query().Get("institute")
		if institute == "" {
			msg := "institute parameter is required"
			log.Error(msg)
			render.JSON(w, r, resp.Error(msg))
			return
		}

		birthdays, err := birthdayGetter.GetTomorrowsBirthdays(ctx, institute)
		if err != nil {
			msg := "failed to get birthdays"
			log.Error(msg, slog.String("institute", institute), sl.Err(err))
			render.JSON(w, r, resp.Error(msg))
			return
		}

		log.Info("birthdays retrieved successfully", slog.Int("count", len(birthdays)))

		render.JSON(w, r, birthdays)

	}
}
