package rest

import (
	"context"

	"encore.dev/beta/errs"
)

type DummyTokenDb struct {
	Tokens map[string]SessionInfo
}

var _ TokenDb = &DummyTokenDb{}

func (d *DummyTokenDb) VerifyToken(ctx context.Context, token string) (SessionInfo, error) {
	if info, ok := d.Tokens[token]; ok {
		return info, nil
	}
	return SessionInfo{}, errs.B().Code(errs.NotFound).Msg("token not found").Err()
}

func (d *DummyTokenDb) Close(ctx context.Context) {}

func CreateFakeDummyTokenDb() *DummyTokenDb {
	return &DummyTokenDb{Tokens: map[string]SessionInfo{
		"token-alice": {CustomerId: "aec31fe6-04b5-4dbf-a024-b5f45db6f633"},
		"token-bob":   {CustomerId: "b59c18af-50be-4f4d-91ad-b25c9c9d0581"},
	}}
}
