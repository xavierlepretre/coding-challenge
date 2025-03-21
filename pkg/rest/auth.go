package rest

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type AuthData struct{}

//encore:authhandler
func (s *BillingService) AuthHandler(ctx context.Context, token string) (auth.UID, *AuthData, error) {
	sessionInfo, err := s.tokenDb.VerifyToken(ctx, token)
	if err != nil {
		return "", nil, errs.WrapCode(err, errs.Unauthenticated, "invalid token")
	}
	return auth.UID(sessionInfo.CustomerId), &AuthData{}, nil
}

type SessionInfo struct {
	CustomerId string
}

type TokenDb interface {
	VerifyToken(ctx context.Context, token string) (SessionInfo, error)
	Close(ctx context.Context)
}

func TokenDbFactory(_ string) (TokenDb, error) {
	return CreateFakeDummyTokenDb(), nil
}
