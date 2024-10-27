package main

import (
	"expvar"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wesleybruno/golang-monolito/internal/auth"
	"github.com/wesleybruno/golang-monolito/internal/db"
	"github.com/wesleybruno/golang-monolito/internal/env"
	"github.com/wesleybruno/golang-monolito/internal/mailer"
	"github.com/wesleybruno/golang-monolito/internal/ratelimiter"
	"github.com/wesleybruno/golang-monolito/internal/store"
	"github.com/wesleybruno/golang-monolito/internal/store/cache"
	"go.uber.org/zap"
)

const version = ""

//	@title			GoSocial API
//	@description	This is a sample of api.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {

	env.LoadConfig()

	cfg := config{
		addr:        env.Config.ApiPort,
		env:         env.Config.Env,
		apiUrl:      env.Config.ApiUrl,
		frontendURL: env.Config.FrontendURL,
		mail: mail{
			exp: time.Hour * 24 * 3, // 3 days
			sendgrid: sendgrid{
				apiKey:    env.Config.SendGridApiKey,
				fromEmail: env.Config.FromEmail,
			},
		},
		dbConfig: dbConfig{
			addr:         fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", env.Config.DbUser, env.Config.DbPassword, env.Config.DbAddress, env.Config.DbName),
			maxOpenConns: env.Config.MaxOpenConns,
			maxIdleConns: env.Config.MaxIdleConns,
			maxIdleTime:  env.Config.MaxIdleTime,
		},
		cache: redisCfg{
			addr:    env.Config.RedisAddr,
			pwd:     env.Config.RedisPwd,
			db:      0,
			enabled: env.Config.RedisEnabled,
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.Config.AuthBasicUser,
				pass: env.Config.AuthBasicPass,
			},
			token: tokenConfig{
				secret: env.Config.JwtSecret,
				exp:    time.Hour * 24 * 3, // 3 days
				iss:    "goapi",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.Config.RateLimiterRequestCount,
			TimeFrame:           time.Second * 30,
			Enabled:             env.Config.RateLimiterEnabled,
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(cfg.dbConfig.addr, cfg.dbConfig.maxOpenConns, cfg.dbConfig.maxIdleConns, cfg.dbConfig.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	var rdb *redis.Client
	if cfg.cache.enabled {
		rdb = cache.NewRedisClient(cfg.cache.addr, cfg.cache.pwd, cfg.cache.db)
		logger.Info("redis database connection established")
	}

	store := store.NewStorage(db)

	cacheStore := cache.NewRedisStorage(rdb)

	mailer := mailer.NewSendGrid(cfg.mail.sendgrid.apiKey, cfg.mail.sendgrid.fromEmail)

	jwtAuthenticator := auth.NewJwtAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	app := &application{
		config:      cfg,
		store:       store,
		cache:       cacheStore,
		logger:      logger,
		mailer:      mailer,
		auth:        jwtAuthenticator,
		rateLimiter: rateLimiter,
	}

	// Metrics collected
	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()

	logger.Fatal(app.run(mux))

}
