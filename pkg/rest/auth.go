package rest

import (
	"coding-challenge/pkg/model"
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
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

func getAuthenticatedCustomerId() (*model.CustomerId, error) {
	// // Use this hack while encore does not return UID when unit testing auth end points.
	// customerId := model.CustomerId("aec31fe6-04b5-4dbf-a024-b5f45db6f633")
	// return &customerId, nil
	authId, ok := auth.UserID()
	if !ok {
		rlog.Error("failed to get user id", ok)
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "failed to get user id",
		}
	}
	customerId := model.CustomerId(authId)
	return &customerId, nil
}
