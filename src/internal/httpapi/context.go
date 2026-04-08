package httpapi

import (
	"context"

	"github.com/MichaelJ43/adaptive-pipe/internal/auth"
)

type ctxKey int

const claimsKey ctxKey = 1

func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*auth.Claims)
	return c, ok
}
