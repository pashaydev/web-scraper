package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const UserClaimsKey contextKey = "userClaims"

func setUserClaimsToContext(ctx context.Context, claims jwt.MapClaims) context.Context {
	return context.WithValue(ctx, UserClaimsKey, claims)
}

func GetUserClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(jwt.MapClaims)
	return claims, ok
}
