package middlewares

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt"
)

const ID_JWT_TOKEN_KEY = "id"

func CreateToken(id string, jwtSecret string) (string, error) {
	payload := jwt.MapClaims{
		ID_JWT_TOKEN_KEY: id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return token.SignedString([]byte(jwtSecret))
}

func UnmarshalToken(jwtToken string, jwtSecret string) (string, error) {
	token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	id, ok := mapClaims[ID_JWT_TOKEN_KEY].(string)
	if !ok {
		return "", err
	}

	return id, nil
}

func CreateValidateAuth(jwtSecret string, api huma.API, idGetter func(huma.Context) string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		logger := GetLoggerFromContext(ctx.Context())
		token := ctx.Header("Authorization")
		if token == "" {
			logger.Info("Rejected request, no authorization header")
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"Failed to authorize, invalid authorization header",
			)
			return
		}

		expected := idGetter(ctx)
		actual, err := UnmarshalToken(token, jwtSecret)
		if err != nil {
			logger.Info("Rejected request, failed to parse jwt token")
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"Failed to authorize, invalid authorization header, failed to parse",
			)
			return
		}
		logger.
			Debug(fmt.Sprintf("Got tokens (expected, actual): (%q, %q)",
				expected,
				actual,
			))

		if expected != actual {
			logger.Info("Rejected request, invalid token")
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"Failed to authorize, invalid authorization header",
			)
			return
		}

		next(ctx)
	}
}
