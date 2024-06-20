package handlers

import (
	"URLShortner/cmd/shortener/middlewares"
	"URLShortner/internal/db"
	"URLShortner/pkg"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/danielgtaylor/huma/v2"
)

type createInput struct {
	Body struct {
		URL     string  `json:"url" example:"http://google.com"`
		Webhook *string `json:"webhook" required:"false" example:"http://localhost:8001/verify"`
	}
}

type CreateOutputBody struct {
	Key          string `json:"key"`
	URLShortened string `json:"url"`
}

type createOutput struct {
	Body    CreateOutputBody
	JWTAuth string `header:"Authorization"`
}

type CreateHandler struct {
	DB         db.URLDatabase
	FollowLink string
	JWTSecret  string
}

func (i *createInput) Resolve(ctx huma.Context) []error {
	var ers []error

	_, err := url.Parse(i.Body.URL)
	if err != nil {
		ers = append(ers, &huma.ErrorDetail{
			Location: "path.URL",
			Message:  "URL is in incorrect format",
			Value:    err.Error(),
		})
	}

	return ers
}

func (h *CreateHandler) GetOperation() huma.Operation {
	return huma.Operation{
		OperationID:   "shortenedCreate",
		Method:        http.MethodPost,
		Path:          "/create",
		Summary:       "Create shortened url",
		DefaultStatus: http.StatusOK,
	}
}

func (h *CreateHandler) Handle(ctx context.Context, body *createInput) (*createOutput, error) {
	logger := middlewares.GetLoggerFromContext(ctx)

	key, err := h.DB.GetFreeKey()

	if err != nil {
		return nil, fmt.Errorf("no free avaliable keys")
	}

	err = h.DB.Create(pkg.URLShortened{
		Key:     key,
		Origin:  body.Body.URL,
		WebHook: body.Body.Webhook,
	})
	if err != nil {
		return nil, err
	}

	jwtToken, err := middlewares.CreateToken(key, h.JWTSecret)
	if err != nil {
		return nil, err
	}

	logger.Debug(fmt.Sprintf("Created new key %q with jwt %q", key, jwtToken))
	return &createOutput{
		Body: CreateOutputBody{
			Key:          key,
			URLShortened: fmt.Sprintf("%s/%s", h.FollowLink, key),
		},
		JWTAuth: jwtToken,
	}, nil
}
