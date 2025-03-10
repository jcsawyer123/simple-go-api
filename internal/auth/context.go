package auth

import (
	"context"
)

type ctxKey string

const (
	// TokenCtxKey is the context key for the authentication token
	TokenCtxKey ctxKey = "auth-token"
)

// WithToken adds a token to the context
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenCtxKey, token)
}

// TokenFromContext extracts a token from the context
func TokenFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	token, ok := ctx.Value(TokenCtxKey).(string)
	return token, ok
}
