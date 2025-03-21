package rest

import (
	"context"
	"testing"

	"encore.dev/beta/auth"
	"github.com/stretchr/testify/assert"
)

func TestDummyAuthHandler_Alice(t *testing.T) {
	s := BillingService{tokenDb: CreateFakeDummyTokenDb()}
	uid, data, err := s.AuthHandler(context.Background(), "token-alice")
	assert.Equal(t, auth.UID("aec31fe6-04b5-4dbf-a024-b5f45db6f633"), uid)
	assert.Equal(t, &AuthData{}, data)
	assert.NoError(t, err)
}

func TestDummyAuthHandler_Fail(t *testing.T) {
	s := BillingService{tokenDb: CreateFakeDummyTokenDb()}
	uid, data, err := s.AuthHandler(context.Background(), "token-will")
	assert.Equal(t, auth.UID(""), uid)
	assert.Nil(t, data)
	assert.Error(t, err)
}
