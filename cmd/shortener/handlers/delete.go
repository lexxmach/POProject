package handlers

import (
	"URLShortner/cmd/shortener/middlewares"
	"URLShortner/internal/db"
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type deleteInput struct {
	Key         string `json:"key" path:"key"`
	OnlyWebhook bool   `param:"webhook" required:"false" description:"If present and true, will delete only webhook"`

	Token string `header:"Authorization" required:"true"`
}

type DeleteHandler struct {
	DB        db.URLDatabase
	JWTSecret string
}

func (h *DeleteHandler) GetOperationAllDelete(api huma.API) huma.Operation {
	return huma.Operation{
		OperationID: "shortenedDelete",
		Method:      http.MethodDelete,
		Path:        "/delete/{key}",
		Summary:     "Delete shortened url key",
		Middlewares: []func(ctx huma.Context, next func(huma.Context)){
			middlewares.CreateValidateAuth(h.JWTSecret, api, func(ctx huma.Context) string {
				return ctx.Param("key")
			}),
		},
		DefaultStatus: http.StatusOK,
	}
}

func (h *DeleteHandler) GetOperationWebhookDelete(api huma.API) huma.Operation {
	return huma.Operation{
		OperationID: "shortenedDeleteWebhook",
		Method:      http.MethodDelete,
		Path:        "/delete/webhook/{key}",
		Summary:     "Delete webhook by url key",
		Middlewares: []func(ctx huma.Context, next func(huma.Context)){
			middlewares.CreateValidateAuth(h.JWTSecret, api, func(ctx huma.Context) string {
				return ctx.Param("key")
			}),
		},
		DefaultStatus: http.StatusOK,
	}
}

func (h *DeleteHandler) handle(ctx context.Context, body *deleteInput) (*struct{}, error) {
	logger := middlewares.GetLoggerFromContext(ctx)

	var err error
	if body.OnlyWebhook {
		err = h.DB.DeleteWebhook(body.Key)
		logger.Debug(fmt.Sprintf("Deleted webhook key %q", body.Key))
	} else {
		err = h.DB.Delete(body.Key)
		logger.Debug(fmt.Sprintf("Deleted key %q", body.Key))
	}
	if err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}

func (h *DeleteHandler) HandleAllDelete(ctx context.Context, body *deleteInput) (*struct{}, error) {
	body.OnlyWebhook = false
	return h.handle(ctx, body)
}

func (h *DeleteHandler) HandleWebhookDelete(ctx context.Context, body *deleteInput) (*struct{}, error) {
	body.OnlyWebhook = true
	return h.handle(ctx, body)
}
