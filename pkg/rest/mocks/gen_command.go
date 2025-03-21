package mocks

import (
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/rest"

	"go.temporal.io/sdk/client"
)

//go:generate mockgen -destination=mock_client.go -package=mocks go.temporal.io/sdk/client Client
var _ client.Client = &MockClient{}

// Then, in mock_client.go, replace:
// internal "go.temporal.io/sdk/internal"
// with
// internal "go.temporal.io/sdk/client"
//

//go:generate mockgen -destination=mock_tokendb.go -package=mocks -source=../auth.go
var _ rest.TokenDb = &MockTokenDb{}

//go:generate mockgen -destination=mock_id.go -package=mocks -source=../../model/id.go
var _ model.BillIdGenerator = &MockBillIdGenerator{}
