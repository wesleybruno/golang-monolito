package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/wesleybruno/golang-monolito/docs"
	"github.com/wesleybruno/golang-monolito/internal/auth"
	"github.com/wesleybruno/golang-monolito/internal/mailer"
	"github.com/wesleybruno/golang-monolito/internal/store"
	"github.com/wesleybruno/golang-monolito/internal/store/cache"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
	mailer mailer.Client
	auth   auth.Authenticator
	cache  cache.Storage
}

type config struct {
	dbConfig    dbConfig
	addr        string
	env         string
	apiUrl      string
	frontendURL string
	mail        mail
	auth        authConfig
	cache       redisCfg
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type redisCfg struct {
	addr    string
	pwd     string
	db      int
	enabled bool
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type mail struct {
	exp      time.Duration
	sendgrid sendgrid
}

type sendgrid struct {
	apiKey    string
	fromEmail string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsUrl),
		))

		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckerHandler)

		r.Route("/post", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.CheckPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.CheckPostOwnership("admin", app.deletePostHandler))
			})
		})

		r.Route("/user", func(r chi.Router) {

			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})

		})
		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r

}

func (app *application) run(mux http.Handler) error {
	//Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiUrl
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("server started at port", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}
