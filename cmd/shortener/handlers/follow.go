package handlers

import (
	"URLShortner/cmd/shortener/middlewares"
	"URLShortner/internal/db"
	"URLShortner/pkg"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type followInput struct {
	Key     string `path:"key" json:"key" required:"true"`
	Headers http.Header
}

type followOutput struct {
	NewUrl string `header:"Location"`
}

type FollowHandler struct {
	DB db.URLDatabase
}

func (i *followInput) Resolve(ctx huma.Context) []error {
	i.Headers = make(http.Header)
	ctx.EachHeader(func(name, value string) {
		i.Headers.Add(name, value)
	})
	return nil
}

func (h *FollowHandler) GetOperation() huma.Operation {
	return huma.Operation{
		OperationID:   "follow",
		Method:        http.MethodGet,
		Path:          "/follow/{key}",
		Summary:       "Follow shortened url",
		DefaultStatus: http.StatusMovedPermanently,
	}
}

func (h *FollowHandler) Handle(ctx context.Context, body *followInput) (*followOutput, error) {
	logger := middlewares.GetLoggerFromContext(ctx)

	avaliable, err := h.DB.Avaliable(body.Key)
	if avaliable {
		return nil, huma.Error400BadRequest("Key doesnt exist")
	}
	if err != nil {
		return nil, err
	}

	url, err := h.DB.Get(body.Key)
	if err != nil {
		return nil, err
	}

	if url.WebHook != nil {
		resp, err := SendWebhookRequest(ctx, body, *url.WebHook)
		if err != nil {
			logger.Info(fmt.Sprintf("Webhook request to %q failed with: %q", *url.WebHook, err.Error()))

			return nil, huma.Error500InternalServerError("Webhook request failed")
		}

		logger.Info(fmt.Sprintf("Webhook request to %q responded (%t, %q)", *url.WebHook, resp.Pass, resp.Reason))
		if !resp.Pass {
			return nil, huma.Error403Forbidden(resp.Reason)
		}
	}

	return &followOutput{
		NewUrl: url.Origin,
	}, nil
}

func SendWebhookRequest(ctx context.Context, body *followInput, webhookURL string) (*pkg.WebHookResponse, error) {
	req, err := http.NewRequest("GET", webhookURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = body.Headers

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	webhookResponse := &pkg.WebHookResponse{}
	err = json.Unmarshal(respBytes, webhookResponse)
	if err != nil {
		return nil, err
	}

	return webhookResponse, nil
}
