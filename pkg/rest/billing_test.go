package rest_test

import (
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/rest"
	"coding-challenge/pkg/rest/mocks"
	"coding-challenge/pkg/workflow"
	"context"
	"testing"
	"time"

	"encore.dev/beta/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func encodeMockedState(ctrl *gomock.Controller, state workflow.BillingState) *mocks.MockEncodedValue {
	encodedBillingstate := mocks.NewMockEncodedValue(ctrl)
	encodedBillingstate.EXPECT().
		Get(gomock.Any()).
		SetArg(0, state).
		Return(nil)
	return encodedBillingstate
}

func createBasicMocks(ctrl *gomock.Controller, billInfo model.BillInfo) (
	*mocks.MockWorkflowRun,
	*mocks.MockClient,
	*mocks.MockTokenDb,
	*mocks.MockBillIdGenerator,
) {
	worflowRun := mocks.NewMockWorkflowRun(ctrl)
	worflowRun.EXPECT().GetID().Return("mock-wr-id")
	worflowRun.EXPECT().GetRunID().Return("mock-run-id")
	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		ExecuteWorkflow(
			gomock.Any(), gomock.Any(), gomock.Any(),
			billInfo,
			gomock.Any()).
		Return(worflowRun, nil)
	tokenDb := mocks.NewMockTokenDb(ctrl)
	// tokenDb.EXPECT(). // For some reason, unit testing the auth end point does not work as expected.
	// 	VerifyToken(gomock.Any(), gomock.Eq("token-alice")).
	// 	Return(rest.SessionInfo{string(billInfo.Id.CustomerId)}, nil)
	billIdGenerator := mocks.NewMockBillIdGenerator(ctrl)
	billIdGenerator.EXPECT().
		New().
		Return(billInfo.Id.Id)
	return worflowRun, client, tokenDb, billIdGenerator
}

func addGetExpectations(ctrl *gomock.Controller, client *mocks.MockClient, billingStates ...workflow.BillingState) {
	for _, billingState := range billingStates {
		client.EXPECT().
			QueryWorkflow(
				gomock.Any(), gomock.Any(),
				gomock.Any(),
				workflow.GetPendingBillStateQuery).
			Return(encodeMockedState(ctrl, billingState), nil).
			Times(1)
	}
}

func addCloseExpectations(ctrl *gomock.Controller, client *mocks.MockClient, finalState workflow.BillingState) *mocks.MockWorkflowRun {
	client.EXPECT().SignalWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	workflowRun := mocks.NewMockWorkflowRun(ctrl)
	workflowRun.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		SetArg(1, finalState).
		Return(nil)
	client.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).Return(workflowRun)
	return workflowRun
}

func addAddLineItemExpectations(
	ctrl *gomock.Controller,
	client *mocks.MockClient,
	billIdGenerator *mocks.MockBillIdGenerator,
	lineItem model.BillLineItem,
	updateId string,
	updatedState workflow.BillingState,
) *mocks.MockWorkflowUpdateHandle {
	billIdGenerator.EXPECT().New().Return(updateId)
	billIdGenerator.EXPECT().New().Return(lineItem.Id.Id)
	updateHandle := mocks.NewMockWorkflowUpdateHandle(ctrl)
	updateHandle.EXPECT().Get(gomock.Any(), gomock.Any()).SetArg(1, updatedState).Return(nil)
	client.EXPECT().UpdateWorkflow(gomock.Any(), gomock.Any()).Return(updateHandle, nil)
	return updateHandle
}

func TestOpenNewBill(t *testing.T) {
	// Arrange
	newBill := model.BillInfo{
		Id: model.BillId{
			CustomerId: model.CustomerId("aec31fe6-04b5-4dbf-a024-b5f45db6f633"),
			Id:         "fc03932f-2b53-4d07-ad55-24fc7d85e277",
		},
		CurrencyCode: "USD",
		Status:       model.Open}
	authedContext := auth.WithContext(context.Background(), auth.UID(newBill.Id.CustomerId), &rest.AuthData{})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	_, client, tokenDb, billIdGenerator := createBasicMocks(ctrl, newBill)
	initialBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 0,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 0, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	addGetExpectations(ctrl, client, initialBillingState)
	s := rest.NewBillingService(client, rest.TokenDb(tokenDb), billIdGenerator)

	// Act
	resp, err := s.OpenNewBill(authedContext, &rest.OpenNewBillRequest{
		CurrencyCode: "USD",
		CloseTime:    time.Now().Add(time.Minute),
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t,
		&rest.OpenNewBillResponse{Id: newBill.Id.Id},
		resp)
}

func TestGetBill(t *testing.T) {
	// Arrange
	newBill := model.BillInfo{
		Id: model.BillId{
			CustomerId: model.CustomerId("aec31fe6-04b5-4dbf-a024-b5f45db6f633"),
			Id:         "fc03932f-2b53-4d07-ad55-24fc7d85e277",
		},
		CurrencyCode: "USD",
		Status:       model.Open}
	authedContext := auth.WithContext(context.Background(), auth.UID(newBill.Id.CustomerId), &rest.AuthData{})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	_, client, tokenDb, billIdGenerator := createBasicMocks(ctrl, newBill)
	initialBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 0,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 0, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	addGetExpectations(ctrl, client, initialBillingState, initialBillingState)
	s := rest.NewBillingService(client, rest.TokenDb(tokenDb), billIdGenerator)
	_, err := s.OpenNewBill(authedContext, &rest.OpenNewBillRequest{
		CurrencyCode: "USD",
		CloseTime:    time.Now().Add(time.Minute),
	})
	assert.NoError(t, err)

	// Act
	resp, err := s.GetBill(authedContext, newBill.Id.Id, &rest.GetBillRequest{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t,
		&rest.GetBillResponse{
			Id:            newBill.Id.Id,
			CurrencyCode:  newBill.CurrencyCode,
			Status:        model.Open,
			LineItemCount: 0,
			TotalOk:       "y",
			Total:         0,
		},
		resp)
}

func TestCloseBill(t *testing.T) {
	// Arrange
	newBill := model.BillInfo{
		Id: model.BillId{
			CustomerId: model.CustomerId("aec31fe6-04b5-4dbf-a024-b5f45db6f633"),
			Id:         "fc03932f-2b53-4d07-ad55-24fc7d85e277",
		},
		CurrencyCode: "USD",
		Status:       model.Open}
	authedContext := auth.WithContext(context.Background(), auth.UID(newBill.Id.CustomerId), &rest.AuthData{})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	_, client, tokenDb, billIdGenerator := createBasicMocks(ctrl, newBill)
	initialBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 0,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 0, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	addGetExpectations(ctrl, client, initialBillingState)
	finalBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 0,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 0, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	_ = addCloseExpectations(ctrl, client, finalBillingState)
	s := rest.NewBillingService(client, rest.TokenDb(tokenDb), billIdGenerator)
	_, err := s.OpenNewBill(authedContext, &rest.OpenNewBillRequest{
		CurrencyCode: "USD",
		CloseTime:    time.Now().Add(time.Minute),
	})
	assert.NoError(t, err)

	// Act
	resp, err := s.CloseBill(authedContext, newBill.Id.Id, &rest.CloseBillRequest{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t,
		&rest.CloseBillResponse{
			CurrencyCode:  "USD",
			LineItemCount: 0,
			TotalOk:       "y",
			Total:         0,
		},
		resp)
}

func TestAddLineItem(t *testing.T) {
	// Arrange
	newBill := model.BillInfo{
		Id: model.BillId{
			CustomerId: model.CustomerId("aec31fe6-04b5-4dbf-a024-b5f45db6f633"),
			Id:         "fc03932f-2b53-4d07-ad55-24fc7d85e277",
		},
		CurrencyCode: "USD",
		Status:       model.Open}
	authedContext := auth.WithContext(context.Background(), auth.UID(newBill.Id.CustomerId), &rest.AuthData{})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	_, client, tokenDb, billIdGenerator := createBasicMocks(ctrl, newBill)
	initialBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 0,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 0, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	addGetExpectations(ctrl, client, initialBillingState)
	lineItem := model.BillLineItem{
		Id:          model.BillLineItemId{BillId: newBill.Id, Id: "a579a2e5-9c31-473e-94ed-577c7cd14acd"},
		Description: "Matchbox",
		Amount:      model.Amount{Number: 100, CurrencyCode: "USD"},
	}
	updatedBillingState := workflow.BillingState{
		BillInfo:          newBill,
		BillLineItemCount: 1,
		Total: workflow.TotalAmount{
			Total: model.Amount{Number: 100, CurrencyCode: newBill.CurrencyCode},
			Ok:    true,
		},
	}
	s := rest.NewBillingService(client, rest.TokenDb(tokenDb), billIdGenerator)
	_, err := s.OpenNewBill(authedContext, &rest.OpenNewBillRequest{
		CurrencyCode: "USD",
		CloseTime:    time.Now().Add(time.Minute),
	})
	assert.NoError(t, err)
	_ = addAddLineItemExpectations(ctrl, client, billIdGenerator, lineItem, "a8f2784e-a7e6-45b6-ad09-8186422a9261", updatedBillingState)

	// Act
	resp, err := s.AddBillLineItem(authedContext, newBill.Id.Id, &rest.AddBillLineItemRequest{
		Description:  "Matchbox",
		Amount:       100,
		CurrencyCode: "USD",
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t,
		&rest.AddBillLineItemResponse{
			Id:            lineItem.Id.Id,
			CurrencyCode:  "USD",
			LineItemCount: 1,
			TotalOk:       "y",
			Total:         100,
		},
		resp)
}
