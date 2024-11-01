package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/Talonmortem/Rest-API-service/internal/lib/api/response"
	"github.com/Talonmortem/Rest-API-service/internal/lib/logger/sl"
	"github.com/Talonmortem/Rest-API-service/internal/lib/random"
	"github.com/Talonmortem/Rest-API-service/internal/storage"

	//"github.com/Talonmortem/Rest-API-service/internal/storage/sqlite"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSAve string, alias string) (int64, error)
}

//go:generate go run github.com/vektra/mockery/v2 --name=URLSaver

const aliasLength = 6

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("failed to validate request body", sl.Err(err))
			render.JSON(w, r, resp.ValidateError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
			/*s := &sqlite.Storage{}
			err := s.CheckAlias(alias)
			for err != nil {
				//generate a new alias
				alias = random.NewRandomString(aliasLength)
				err = s.CheckAlias(alias)
			}*/
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url already exists"))
			return

		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add url"))
			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
