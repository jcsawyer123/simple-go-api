package auth

import "context"

type ctxKey string

const (
	TokenCtxKey ctxKey = "token" // Note: Exported (capitalized)
)

func TokenFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	token, ok := ctx.Value(TokenCtxKey).(string)
	return token, ok
}
