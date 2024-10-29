package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Talonmortem/Rest-API-service/internal/lib/api/response"
	"github.com/Talonmortem/Rest-API-service/internal/lib/logger/sl"
	"github.com/Talonmortem/Rest-API-service/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2 --name=URLGetter

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias is empty")
			render.JSON(w, r, resp.Error("invalid data: alias is empty"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", sl.Err(err))
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to get url"))
			return
		}

		log.Info("got url", slog.String("url", resURL))
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
