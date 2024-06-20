package middlewares

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type CONTEXT_LOG_TYPE string

const CONTEXT_LOG_KEY CONTEXT_LOG_TYPE = "logger"

func GetLoggerFromContext(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(CONTEXT_LOG_KEY)

	if logger != nil {
		casted, ok := logger.(*logrus.Entry)

		if !ok {
			return nil
		}
		return casted
	}
	return nil
}

func SetLoggerToContext(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, CONTEXT_LOG_KEY, entry)
}

func RequestIDMiddleware(ctx huma.Context, next func(huma.Context)) {
	requestID := uuid.New().String()

	logger := GetLoggerFromContext(ctx.Context()).WithField("request-id", requestID)

	logger.Debug("Got request")
	next(huma.WithContext(ctx, SetLoggerToContext(ctx.Context(), logger)))
}

func CreateLogMiddleware(logger *logrus.Logger) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		url := ctx.URL()
		logger := logger.WithContext(ctx.Context()).WithField("method", url.String())

		next(huma.WithContext(ctx, SetLoggerToContext(ctx.Context(), logger)))
	}
}
