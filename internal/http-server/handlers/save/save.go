package save

import (
	"errors"
	resp "github.com/AlexLex13/Infinity/internal/lib/api/response"
	"github.com/AlexLex13/Infinity/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

type URLSaver interface {
	SaveURL(URL, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		// Добавляем к текущму объекту логгера поля op и request_id
		// Они могут очень упростить нам жизнь в будущем
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Создаем объект запроса и анмаршаллим в него запрос
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request")) // <----

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request")) // <----

			return
		}

		// Лучше больше логов, чем меньше - лишнее мы легко сможем почистить,
		// при необходимости. А вот недостающую информацию мы уже не получим.
		log.Info("request body decoded", slog.Any("req", req))

		// ...
	}
}
