package workflow_test

import (
	"coding-challenge/pkg/activity"
	"coding-challenge/pkg/model"
	"coding-challenge/pkg/workflow"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type BillingWorkflowUnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func TestBillingWorkflowUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BillingWorkflowUnitTestSuite))
}

func (s *BillingWorkflowUnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *BillingWorkflowUnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *BillingWorkflowUnitTestSuite) defaultBillAndItems() (billInfo model.BillInfo, lineItem1 model.BillLineItem, lineItem2 model.BillLineItem) {
	billInfo = model.BillInfo{
		Id:           model.BillId{CustomerId: "alice", Id: "ca06186a-1f96-4398-9244-fbddf4ef2642"},
		CurrencyCode: "USD",
		Status:       model.Open,
	}
	amount1, err := model.NewAmountFromInt64(100, "USD")
	s.NoError(err)
	lineItem1 = model.BillLineItem{
		Id:          model.BillLineItemId{BillId: billInfo.Id, Id: "5a61aae5-e120-4ddb-a15a-34cdfa74a1b6"},
		Description: "Matchbox",
		Amount:      amount1,
	}
	amount2, err := model.NewAmountFromInt64(200, "USD")
	s.NoError(err)
	lineItem2 = model.BillLineItem{
		Id:          model.BillLineItemId{BillId: billInfo.Id, Id: "9497a0e4-f59d-4382-a978-6728ab62e7f5"},
		Description: "Candle",
		Amount:      amount2,
	}
	return billInfo, lineItem1, lineItem2
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_Fails_NegativeDuration() {
	// Arrange
	billInfo, _, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil).Never()
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Never()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil).Never()

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Hour*-1)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	e := s.env.GetWorkflowError()
	s.Error(e, "duration is negative")
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	s.Equal(workflow.BillingState{}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseAtMaturity_WithoutItems() {
	// Arrange
	billInfo, _, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Never()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Hour*24*30)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 0,
		Total:             model.TotalAmount{Total: model.Amount{Number: 0, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseEarly_WithoutItems() {
	// Arrange
	billInfo, _, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Never()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		message := "Close bill"
		s.env.SignalWorkflow(workflow.CloseBillEarlySignal, &message)
	}, 2*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Hour*24*30)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 0,
		Total:             model.TotalAmount{Total: model.Amount{Number: 0, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseEarly_WithFailedItem() {
	// Arrange
	billInfo, lineItem, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(0), errors.New("Fake error")).Times(10) // 10 attempts seem to be made by default
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.Error(err) },
				OnReject:   func(err error) { s.ErrorIs(err, errors.New("Fake error")) },
			},
			lineItem)
	}, 1*time.Second)
	s.env.RegisterDelayedCallback(func() {
		message := "Close bill"
		s.env.SignalWorkflow(workflow.CloseBillEarlySignal, &message)
	}, time.Hour) // An hour to give time for the 10 attempts

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Hour) // An hour to give time for the 10 attempts

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(0), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 0,
		Total:             model.TotalAmount{Total: model.Amount{Number: 0, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseEarly_With1Item() {
	// Arrange
	billInfo, lineItem, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil)
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnComplete: func(result interface{}, err error) {
					s.NoError(err)
					intermediateState := result.(workflow.BillingState)
					s.Equal(workflow.BillingState{
						BillInfo:          billInfo,
						BillLineItemCount: 1,
						Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
					}, intermediateState)
				},
				OnReject: func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem)
	}, 1*time.Second)
	s.env.RegisterDelayedCallback(func() {
		message := "Close bill"
		s.env.SignalWorkflow(workflow.CloseBillEarlySignal, &message)
	}, 2*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(1), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 1,
		Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseAtMaturity_With2ItemsTogether() {
	// Arrange
	billInfo, lineItem1, lineItem2 := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Twice()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		nextCount := 1
		updateCallback := testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnComplete: func(result interface{}, err error) {
				s.NoError(err)
				intermediateState := result.(workflow.BillingState)
				s.Equal(uint64(nextCount), intermediateState.BillLineItemCount)
				nextCount++ // There appears to be an uncertainty in the order of the updates being called.
			},
			OnReject: func(err error) { s.FailNow("Should not reach here") },
		}
		s.env.UpdateWorkflow(workflow.AddBillLineItemUpdate, "1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4", &updateCallback, lineItem1)
		s.env.UpdateWorkflow(workflow.AddBillLineItemUpdate, "ed20aa79-5ddc-4510-a5a3-cda08372e273", &updateCallback, lineItem2)
	}, 1*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(2), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 2,
		Total:             model.TotalAmount{Total: model.Amount{Number: 300, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseAtMaturity_With2ItemsSpaced() {
	// Arrange
	billInfo, lineItem1, lineItem2 := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Twice()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.NoError(err) },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem1)
	}, 1*time.Second)
	var intermediateState workflow.BillingState
	s.env.RegisterDelayedCallback(func() {
		encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
		s.NoError(err)
		encodedState.Get(&intermediateState)
		s.Equal(workflow.BillingState{
			BillInfo:          billInfo,
			BillLineItemCount: 1,
			Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
		}, intermediateState)
	}, 3*time.Second)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"ed20aa79-5ddc-4510-a5a3-cda08372e273",
			&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnComplete: func(result interface{}, err error) {
					s.NoError(err)
					intermediateState := result.(workflow.BillingState)
					s.Equal(workflow.BillingState{
						BillInfo:          billInfo,
						BillLineItemCount: 2,
						Total:             model.TotalAmount{Total: model.Amount{Number: 300, CurrencyCode: "USD"}, Ok: true},
					}, intermediateState)
				},
				OnReject: func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem2)
	}, 5*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(2), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 2,
		Total:             model.TotalAmount{Total: model.Amount{Number: 300, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

// This is more of a test to learn about the behaviour of Temporal.
func (s *BillingWorkflowUnitTestSuite) Test_Workflow_AddSameUpdateId_OnlyFirstRecorded() {
	// Arrange
	billInfo, lineItem1, lineItem2 := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(func(lineItem model.BillLineItem, _ model.TotalAmount) (uint64, error) {
		// Only the first will be called
		s.Equal(lineItem1.Id.Id, lineItem.Id.Id)
		return uint64(1), nil
	})
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.NoError(err) },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem1)
	}, 1*time.Second)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			// Same update id
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnComplete: func(result interface{}, err error) {
					s.NoError(err)
					intermediateState := result.(workflow.BillingState)
					// Still 1 and 100
					s.Equal(workflow.BillingState{
						BillInfo:          billInfo,
						BillLineItemCount: 1,
						Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
					}, intermediateState)
				},
				OnReject: func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem2)
	}, 2*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(1), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 1,
		Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_AddSameItemId_OnlyFirstRecorded() {
	// Arrange
	billInfo, lineItem1, _ := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	calledTimes := uint64(0)
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(func(lineItem model.BillLineItem, _ model.TotalAmount) (uint64, error) {
		// Mimic idempotence
		updateCount := 1 - calledTimes
		calledTimes = 1
		return updateCount, nil
	}).Twice()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.NoError(err) },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem1)
	}, 1*time.Second)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"ed20aa79-5ddc-4510-a5a3-cda08372e273",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.NoError(err) },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			// Same line item
			lineItem1)
	}, 2*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(1), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 1,
		Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CloseAtMaturity_With2Items_TotalOverflow() {
	// Arrange
	billInfo, lineItem1, lineItem2 := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	// Adding to it can only overflow
	lineItem1.Amount.Number = math.MaxInt64
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil).Twice()
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnComplete: func(result interface{}, err error) {
					s.NoError(err)
					intermediateState := result.(workflow.BillingState)
					s.Equal(workflow.BillingState{
						BillInfo:          billInfo,
						BillLineItemCount: 1,
						Total:             model.TotalAmount{Total: model.Amount{Number: math.MaxInt64, CurrencyCode: "USD"}, Ok: true},
					}, intermediateState)
				},
				OnReject: func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem1)
	}, 1*time.Second)
	var intermediateState workflow.BillingState
	s.env.RegisterDelayedCallback(func() {
		encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
		s.NoError(err)
		encodedState.Get(&intermediateState)
		s.Equal(uint64(1), intermediateState.BillLineItemCount)
		s.True(intermediateState.Total.Ok)
		s.Equal(model.Amount{Number: math.MaxInt64, CurrencyCode: "USD"}, intermediateState.Total.Total)
	}, 3*time.Second)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"ed20aa79-5ddc-4510-a5a3-cda08372e273",
			&testsuite.TestUpdateCallback{
				OnAccept: func() {},
				OnComplete: func(result interface{}, err error) {
					s.NoError(err)
					intermediateState := result.(workflow.BillingState)
					s.Equal(workflow.BillingState{
						BillInfo:          billInfo,
						BillLineItemCount: 2,
						Total:             model.TotalAmount{Total: model.Amount{}, Ok: false},
					}, intermediateState)
				},
				OnReject: func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem2)
	}, 5*time.Second)

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, time.Minute)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(2), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 2,
		Total:             model.TotalAmount{Total: model.Amount{}, Ok: false},
	}, result)
}

func (s *BillingWorkflowUnitTestSuite) Test_Workflow_CannotAddItemAfterClose() {
	// Arrange
	billInfo, lineItem1, lineItem2 := s.defaultBillAndItems()
	dummyActivityHost := activity.DummyActivityHost{}
	s.env.OnActivity(dummyActivityHost.CreateBillIfNotExistActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.OnActivity(
		dummyActivityHost.AddBillLineItemIfNotExistActivity,
		mock.AnythingOfType("BillLineItem"),
		mock.AnythingOfType("TotalAmount"),
	).Return(uint64(1), nil)
	s.env.OnActivity(dummyActivityHost.CloseBillActivity, mock.AnythingOfType("BillInfo")).Return(uint64(1), nil)
	s.env.RegisterDelayedCallback(func() {
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"1d1209d3-e60d-4d9c-ae7c-3282f8f5c9b4",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() {},
				OnComplete: func(result interface{}, err error) { s.NoError(err) },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem1)
	}, 1*time.Second)
	s.env.RegisterDelayedCallback(func() {
		s.True(s.env.IsWorkflowCompleted())
		s.env.UpdateWorkflow(
			workflow.AddBillLineItemUpdate,
			"ed20aa79-5ddc-4510-a5a3-cda08372e273",
			&testsuite.TestUpdateCallback{
				OnAccept:   func() { s.FailNow("Should not reach here") },
				OnComplete: func(result interface{}, err error) { s.FailNow("Should not reach here") },
				OnReject:   func(err error) { s.FailNow("Should not reach here") },
			},
			lineItem2)
	}, 5*time.Second) // After maturity

	// Act
	s.env.ExecuteWorkflow(workflow.BillingWorkflow, billInfo, 2*time.Second)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.NoError(err)
	encodedState, err := s.env.QueryWorkflow(workflow.GetPendingBillStateQuery)
	s.NoError(err)
	var receivedState workflow.BillingState
	encodedState.Get(&receivedState)
	s.Equal(uint64(1), receivedState.BillLineItemCount)
	var result workflow.BillingState
	s.env.GetWorkflowResult(&result)
	billInfo.Status = model.Closed
	s.Equal(workflow.BillingState{
		BillInfo:          billInfo,
		BillLineItemCount: 1,
		Total:             model.TotalAmount{Total: model.Amount{Number: 100, CurrencyCode: "USD"}, Ok: true},
	}, result)
}
