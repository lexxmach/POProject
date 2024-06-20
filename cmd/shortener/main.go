package main

import (
	"URLShortner/cmd/shortener/handlers"
	"URLShortner/cmd/shortener/middlewares"
	"URLShortner/internal"
	"URLShortner/internal/db"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gorm.io/driver/postgres"
)

func setupRoutes(config *internal.Config, database db.URLDatabase, logger *logrus.Logger, api huma.API) {
	api.UseMiddleware(
		middlewares.CreateLogMiddleware(logger),
		middlewares.RequestIDMiddleware,
	)

	// CreateDB
	{
		createHandler := handlers.CreateHandler{
			DB:         database,
			FollowLink: config.FollowLink,
			JWTSecret:  config.JWTSecret,
		}
		huma.Register(api, createHandler.GetOperation(), createHandler.Handle)

		followHandler := handlers.FollowHandler{
			DB: database,
		}
		huma.Register(api, followHandler.GetOperation(), followHandler.Handle)

		deleteHandler := handlers.DeleteHandler{
			DB:        database,
			JWTSecret: config.JWTSecret,
		}
		huma.Register(api, deleteHandler.GetOperationAllDelete(api), deleteHandler.HandleAllDelete)
		huma.Register(api, deleteHandler.GetOperationWebhookDelete(api), deleteHandler.HandleWebhookDelete)
	}
}

func main() {
	config := GetConfig()

	logger := logrus.New()
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig(config.Title, config.Version))

	switch config.LogLevel {
	case "DEBUG":
		logger.SetLevel(logrus.DebugLevel)
	case "INFO":
		logger.SetLevel(logrus.InfoLevel)
	default:
		panic(fmt.Errorf("unkown log level: %q", config.LogLevel))
	}

	var database db.URLDatabase
	{
		// TODO(lexmach): make this dynamic config?
		opts := db.KeyGetterOptions{
			MinLength: config.URLShortenedSettins.MinLength,
			MaxLength: config.URLShortenedSettins.MaxLength,
			Retries:   config.URLShortenedSettins.Retries,
			RuneSet:   []rune(config.URLShortenedSettins.Runes),
		}

		if config.DB.Mock {
			database = db.CreateURLMockDB(opts)
		} else {
			dsn := fmt.Sprintf(
				"host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable",
				config.DB.Host,
				config.DB.User,
				config.DB.Password,
				config.DB.Port,
			)

			database = internal.Must(db.CreateURLGormDB(postgres.Open(dsn), opts))
		}
	}

	setupRoutes(config, database, logger, api)

	logrus.Info("Starting server at port ", config.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router); err != nil {
		panic(err)
	}
}

func GetConfig() *internal.Config {
	configPath := pflag.StringP("config", "c", "", "Path to config")
	pflag.Parse()

	return internal.MustGetConfig(*configPath)
}
