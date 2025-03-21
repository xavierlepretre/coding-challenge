package rest_test

import (
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/rest"
	"coding-challenge/pkg/rest/mocks"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestOpenNewBill(t *testing.T) {
	/// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		ExecuteWorkflow(
			gomock.Any(), gomock.Any(), gomock.Any(),
			model.BillInfo{CurrencyCode: "USD", Status: model.Open},
			time.Minute)
	tokenDb := mocks.NewMockTokenDb(ctrl)
	customerId := "aec31fe6-04b5-4dbf-a024-b5f45db6f633"
	tokenDb.EXPECT().
		VerifyToken(gomock.Any(), gomock.Eq("token-alice")).
		Return(rest.SessionInfo{customerId}, nil)
	billIdGenerator := mocks.NewMockBillIdGenerator(ctrl)
	newBillId := "fc03932f-2b53-4d07-ad55-24fc7d85e277"
	billIdGenerator.EXPECT().
		New().
		Return(newBillId)
	s := rest.NewBillingService(client, rest.TokenDb(tokenDb), billIdGenerator)

	// Act
	resp, err := s.OpenNewBill(context.Background(), &rest.OpenNewBillRequest{
		CurrencyCode: "USD",
		CloseTime:    time.Now().Add(time.Minute),
	})
	assert.NoError(t, err)
	assert.Equal(t, &rest.OpenNewBillResponse{
		Id: newBillId,
	}, resp)
}
