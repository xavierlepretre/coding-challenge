package rest

import (
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/workflow"
	"context"
	"fmt"
	"time"

	"encore.dev"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// Use an environment-specific task queue so we can use the same
// Temporal Cluster for all cloud environments.
var (
	envName           = encore.Meta().Environment.Name
	greetingTaskQueue = envName + "-billing"
	tokenDbType       = envName + "-token-db"
)

//encore:service
type BillingService struct {
	client          client.Client
	tokenDb         TokenDb
	billIdGenerator model.BillIdGenerator
}

func initBillingService() (*BillingService, error) {
	client, err := client.Dial(client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %v", err)
	}
	tokenDb, err := TokenDbFactory(tokenDbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create token db: %v", err)
	}
	billIdGenerator := model.UuidBillIdGenerator{}
	return NewBillingService(client, tokenDb, &billIdGenerator), nil
}

func NewBillingService(client client.Client, tokenDb TokenDb, billIdGenerator model.BillIdGenerator) *BillingService {
	return &BillingService{client, tokenDb, billIdGenerator}
}

func (s *BillingService) Shutdown(force context.Context) {
	s.client.Close()
	s.tokenDb.Close(force)
}

type OpenNewBillRequest struct {
	CurrencyCode model.CurrencyCode `json:"currency_code"`
	CloseTime    time.Time          `json:"close_time"`
}

type OpenNewBillResponse struct {
	Id string `json:"id"`
}

func CreateWorkflowId(billId string) string {
	return fmt.Sprintf("create-bill-%v", billId)
}

//encore:api auth method=POST path=/bills
func (s *BillingService) OpenNewBill(ctx context.Context, openNewBillRequest *OpenNewBillRequest) (*OpenNewBillResponse, error) {
	customerId, ok := auth.UserID()
	if !ok {
		rlog.Error("failed to get user id", ok)
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "failed to get user id",
		}
	}

	billId := s.billIdGenerator.New()
	options := client.StartWorkflowOptions{
		ID:        CreateWorkflowId(billId),
		TaskQueue: greetingTaskQueue,
	}
	billInfo := model.BillInfo{
		Id: model.BillId{
			CustomerId: model.CustomerId(customerId),
			Id:         billId,
		},
		CurrencyCode: openNewBillRequest.CurrencyCode,
		Status:       model.Open,
	}
	duration := time.Until(openNewBillRequest.CloseTime)
	wr, err := s.client.ExecuteWorkflow(ctx, options, workflow.BillingWorkflow, billInfo, duration)
	if err != nil {
		rlog.Error("failed to execute workflow", "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "workflow failed to execute")
	}
	runId := wr.GetRunID()
	rlog.Info("started workflow", "id", wr.GetID(), "run_id", runId)

	// Get the intermediate billing state
	var encodedResult converter.EncodedValue
	for i := 0; i < 10; i++ { // HACK ugly wait for the workflow to reach registration of the query handler
		encodedResult, err = s.client.QueryWorkflow(ctx, options.ID, runId, workflow.GetPendingBillStateQuery)
		if err == nil {
			break
		}
		rlog.Error("failed to query workflow", "attempt", i, "err", err)
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Unavailable, "failed to query intermediate state")
	}
	var currentState workflow.BillingState
	err = encodedResult.Get(&currentState)
	if err != nil {
		rlog.Error("failed to decode intermediate state", "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "failed to decode intermediate state")
	} else if currentState.BillInfo.Id.Id != billId {
		rlog.Error("failed to query correct workflow", "billId", billId, "state id", currentState.BillInfo.Id.Id)
		return nil, errs.WrapCode(err, errs.Internal, "failed to query correct workflow")
	}
	return &OpenNewBillResponse{Id: billId}, nil
}

type GetBillRequest struct {
}

type GetBillResponse struct {
	Id            string             `json:"id"`
	CurrencyCode  model.CurrencyCode `json:"currency_code"`
	Status        model.BillStatus   `json:"status"` // open(0)/closed(1)
	LineItemCount uint64             `json:"line_item_count"`
	TotalOk       string             `json:"total_ok"` // y/n instead of true/false
	Total         int64              `json:"total"`
}

const TotalOkYes = "y"
const TotalOkNo = "n"

func getOkString(amount workflow.TotalAmount) string {
	if amount.Ok {
		return TotalOkYes
	} else {
		return TotalOkNo
	}
}

//encore:api auth method=GET path=/bill/:id
func (s *BillingService) GetBill(ctx context.Context, id string, getBillRequest *GetBillRequest) (*GetBillResponse, error) {
	customerId, ok := auth.UserID()
	if !ok {
		rlog.Error("failed to get user id", ok)
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "failed to get user id",
		}
	}

	encodedResult, err := s.client.QueryWorkflow(ctx, CreateWorkflowId(id), "", workflow.GetPendingBillStateQuery)
	if err != nil {
		rlog.Error("failed to query workflow", "err", err)
		return nil, errs.WrapCode(err, errs.NotFound, "failed to query workflow")
	}
	var currentState workflow.BillingState
	err = encodedResult.Get(&currentState)
	if err != nil {
		rlog.Error("failed to decode intermediate state", "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "failed to decode intermediate state")
	} else if currentState.BillInfo.Id.Id != id {
		rlog.Error("failed to query correct workflow", "id", id, "state id", currentState.BillInfo.Id.Id)
		return nil, errs.WrapCode(err, errs.Internal, "failed to query correct workflow")
	} else if currentState.BillInfo.Id.CustomerId != model.CustomerId(customerId) {
		rlog.Error("failed to query workflow of correct customer", "customerId", customerId, "state customer id", currentState.BillInfo.Id.CustomerId)
		return nil, errs.WrapCode(err, errs.Internal, "failed to query correct workflow")
	}

	return &GetBillResponse{
		Id:            id,
		CurrencyCode:  currentState.BillInfo.CurrencyCode,
		Status:        currentState.BillInfo.Status,
		LineItemCount: currentState.BillLineItemCount,
		TotalOk:       getOkString(currentState.Total),
		Total:         currentState.Total.Total.Number,
	}, nil
}

type CloseBillRequest struct {
}

type CloseBillResponse struct {
	CurrencyCode  model.CurrencyCode `json:"currency_code"`
	LineItemCount uint64             `json:"line_item_count"`
	TotalOk       string             `json:"total_ok"` // y/n instead of true/false
	Total         int64              `json:"total"`
}

//encore:api auth method=PATCH path=/bill/:id/close
func (s *BillingService) CloseBill(ctx context.Context, id string, closeBillRequest *CloseBillRequest) (*CloseBillResponse, error) {
	_, ok := auth.UserID()
	if !ok {
		rlog.Error("failed to get user id", ok)
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "failed to get user id",
		}
	}

	err := s.client.SignalWorkflow(ctx, CreateWorkflowId(id), "", workflow.CloseBillEarlySignal, "API initiated")
	if err != nil {
		rlog.Error("failed to close workflow", "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "workflow failed to close")
	}
	rlog.Info("closed workflow", "id", id)
	wr := s.client.GetWorkflow(ctx, CreateWorkflowId(id), "")
	var finalState workflow.BillingState
	err = wr.Get(ctx, &finalState)
	if err != nil {
		rlog.Error("failed to get workflow final state", "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "failed to get workflow final state")
	}
	return &CloseBillResponse{
		CurrencyCode:  finalState.BillInfo.CurrencyCode,
		LineItemCount: finalState.BillLineItemCount,
		TotalOk:       getOkString(finalState.Total),
		Total:         finalState.Total.Total.Number,
	}, nil
}

type AddBillLineItemRequest struct {
	Description  string             `json:"description"`
	Amount       int64              `json:"amount"`
	CurrencyCode model.CurrencyCode `json:"currency-code"`
}

type AddBillLineItemResponse struct {
	Id            string             `json:"id"`
	CurrencyCode  model.CurrencyCode `json:"currency_code"`
	LineItemCount uint64             `json:"line_item_count"`
	TotalOk       string             `json:"total_ok"` // y/n instead of true/false
	Total         int64              `json:"total"`
}

//encore:api auth method=POST path=/bill/:id/line-items
func (s *BillingService) AddBillLineItem(ctx context.Context, id string, addBillLineItemRequest *AddBillLineItemRequest) (*AddBillLineItemResponse, error) {
	customerId, ok := auth.UserID()
	if !ok {
		rlog.Error("failed to get user id", ok)
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "failed to get user id",
		}
	}

	updateId := s.billIdGenerator.New()
	lineItemId := s.billIdGenerator.New()
	options := client.UpdateWorkflowOptions{
		UpdateID:   updateId,
		WorkflowID: CreateWorkflowId(id),
		UpdateName: workflow.AddBillLineItemUpdate,
		Args: []interface{}{
			model.BillLineItem{
				Id: model.BillLineItemId{
					BillId: model.BillId{CustomerId: model.CustomerId(customerId), Id: id},
					Id:     lineItemId,
				},
				Description: addBillLineItemRequest.Description,
				Amount: model.Amount{
					CurrencyCode: addBillLineItemRequest.CurrencyCode,
					Number:       addBillLineItemRequest.Amount,
				},
			},
		},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	}

	updateHandle, err := s.client.UpdateWorkflow(ctx, options)
	if err != nil {
		rlog.Error("failed to add line item", "billId", id, "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "failed to add line item")
	}
	var updatedState workflow.BillingState
	err = updateHandle.Get(ctx, &updatedState)
	if err != nil {
		rlog.Error("failed to get updated workflow state", "billId", id, "err", err)
		return nil, errs.WrapCode(err, errs.Internal, "failed to get updated workflow state")
	}
	rlog.Info("added line item to workflow", "id", id)
	return &AddBillLineItemResponse{
		Id:            lineItemId,
		CurrencyCode:  updatedState.BillInfo.CurrencyCode,
		LineItemCount: updatedState.BillLineItemCount,
		TotalOk:       getOkString(updatedState.Total),
		Total:         updatedState.Total.Total.Number,
	}, nil
}
