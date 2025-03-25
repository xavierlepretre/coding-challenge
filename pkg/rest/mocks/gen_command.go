package mocks

import (
	"coding-challenge/pkg/db"
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/rest"

	"go.temporal.io/sdk/client"
	converter "go.temporal.io/sdk/converter"
)

//go:generate mockgen -destination=mock_client.go -package=mocks go.temporal.io/sdk/client Client
var _ client.Client = &MockClient{}

//go:generate mockgen -destination=mock_workflow_run.go -package=mocks go.temporal.io/sdk/client WorkflowRun
var _ client.WorkflowRun = &MockWorkflowRun{}

// Then, in mock_client.go, replace:
// internal "go.temporal.io/sdk/internal"
// with
// internal "go.temporal.io/sdk/client"

//go:generate mockgen -destination=mock_workflow_update_handle.go -package=mocks go.temporal.io/sdk/client WorkflowUpdateHandle
var _ client.WorkflowUpdateHandle = &MockWorkflowUpdateHandle{}

//go:generate mockgen -destination=mock_encoded_value.go -package=mocks go.temporal.io/sdk/converter EncodedValue
var _ converter.EncodedValue = &MockEncodedValue{}

//go:generate mockgen -destination=mock_tokendb.go -package=mocks -source=../auth.go
var _ rest.TokenDb = &MockTokenDb{}

//go:generate mockgen -destination=mock_id.go -package=mocks -source=../../model/id.go
var _ model.BillIdGenerator = &MockBillIdGenerator{}

//go:generate mockgen -destination=mock_bill_database.go -package=mocks -source=../../db/database.go
var _ db.BillDatabase = &MockBillDatabase{}
