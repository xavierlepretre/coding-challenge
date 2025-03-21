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
