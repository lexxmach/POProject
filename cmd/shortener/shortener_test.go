package main

import (
	"URLShortner/cmd/shortener/handlers"
	"URLShortner/internal"
	"URLShortner/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	pg "gorm.io/driver/postgres"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ParseBody(t *testing.T, resp *http.Response) []byte {
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	return respBytes
}

func TestMock(t *testing.T) {
	for i, config := range []*internal.Config{
		{
			Title:   "xxx",
			Version: "xxx",
			Port:    888,
			DB: internal.ConfigDB{
				Mock: true,
			},
			URLShortenedSettins: internal.ConfigURLSettings{
				MinLength: 4,
				MaxLength: 16,
				Retries:   5,
				Runes:     "abcdef",
			},
			FollowLink: "localhost:888/follow",
			JWTSecret:  "test",
			LogLevel:   "DEBUG",
		},
		{
			Title:   "xxx",
			Version: "xxx",
			Port:    888,
			DB: internal.ConfigDB{
				Mock:     false,
				Host:     "localhost",
				User:     "postgres",
				Password: "postgres",
				Port:     5432,
			},
			URLShortenedSettins: internal.ConfigURLSettings{
				MinLength: 4,
				MaxLength: 16,
				Retries:   5,
				Runes:     "abcdef",
			},
			FollowLink: "localhost:888/follow",
			JWTSecret:  "test",
			LogLevel:   "DEBUG",
		},
	} {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			var database db.URLDatabase
			opts := db.KeyGetterOptions{
				MinLength: config.URLShortenedSettins.MinLength,
				MaxLength: config.URLShortenedSettins.MaxLength,
				Retries:   config.URLShortenedSettins.Retries,
				RuneSet:   []rune(config.URLShortenedSettins.Runes),
			}

			if config.DB.Mock {
				database = db.CreateURLMockDB(opts)
			} else {
				// FIXME(lexmach): gosql fails, need to setup container (wtf?)
				pgContainer, err := postgres.RunContainer(context.Background(),
					testcontainers.WithImage("postgres:15.3-alpine"),
					postgres.WithDatabase("test-db"),
					postgres.WithUsername("postgres"),
					postgres.WithPassword("postgres"),
					testcontainers.WithWaitStrategy(
						wait.ForLog("database system is ready to accept connections").
							WithOccurrence(2).WithStartupTimeout(5*time.Second)),
				)
				assert.NoError(t, err)

				connStr, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
				assert.NoError(t, err)

				database, err = db.CreateURLGormDB(pg.Open(connStr), opts)
				assert.NoError(t, err)
			}

			_, api := humatest.New(t)

			logger := logrus.New()
			setupRoutes(config, database, logger, api)

			resp := api.Post("/create", map[string]any{
				"url": "http://check.org",
			}).Result()

			respBytes := ParseBody(t, resp)

			body := &handlers.CreateOutputBody{}
			auth := resp.Header.Get("Authorization")

			assert.NoError(t, json.Unmarshal(respBytes, body))
			assert.NotEqual(t, auth, "")

			followPath := fmt.Sprintf("/follow/%s", body.Key)
			assert.True(t, strings.HasSuffix(body.URLShortened, followPath))

			resp = api.Get(fmt.Sprintf("/follow/%s", body.Key)).Result()
			assert.Equal(t, resp.StatusCode, http.StatusMovedPermanently)
			assert.Equal(t, "http://check.org", resp.Header.Get("Location"))

			resp = api.Delete(
				fmt.Sprintf("/delete/%s", body.Key),
				fmt.Sprintf("Authorization: %s", "some.badkey"),
			).Result()
			assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

			resp = api.Delete(
				fmt.Sprintf("/delete/%s", body.Key),
				fmt.Sprintf("Authorization: %s", auth),
			).Result()
			assert.Equal(t, resp.StatusCode, http.StatusOK)

			resp = api.Get(fmt.Sprintf("/follow/%s", body.Key)).Result()
			assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
		})
	}
}
