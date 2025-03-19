package workflow

import (
	"fmt"

	"time"

	"coding-challenge/pkg/activity"
	"coding-challenge/pkg/model"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const AddBillLineItemUpdate = "AddBillLineItem"
const GetPendingBillStateQuery = "GetPendingBillState"
const CloseBillEarlySignal = "CloseBillEarly"

type NegativeDurationError struct {
	Duration time.Duration
}

func (e NegativeDurationError) Error() string {
	return fmt.Sprintf("duration is negative %q", e.Duration)
}

type TotalAmount struct {
	Total model.Amount
	Ok    bool
}

func (total *TotalAmount) Add(amount model.Amount) {
	if total.Ok {
		total.Total, total.Ok = total.Total.Add(amount)
	}
}

type BillingState struct {
	BillInfo          model.BillInfo
	BillLineItemCount uint64
	Total             TotalAmount
}

type billingState struct {
	BillingState
	logger log.Logger
}

func (state *billingState) Clone() BillingState {
	return BillingState{
		BillInfo:          state.BillInfo,
		BillLineItemCount: state.BillLineItemCount,
		Total:             state.Total,
	}
}

func defaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: activity.DefaultActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    10,
		},
	}
}

func (state *billingState) createBillIfNotExistSyncActivity(ctx workflow.Context) (uint64, error) {
	state.logger.Info("Creating bill if it does not exist", "Bill", state.BillInfo)
	ctxWithOptions := workflow.WithActivityOptions(ctx, defaultActivityOptions())
	var updateCount uint64
	e := workflow.ExecuteActivity(
		ctxWithOptions,
		activity.CreateBillIfNotExistActivity,
		state.BillInfo, // Potential argument error not caught at compile time?
	).Get(ctxWithOptions, &updateCount)
	return updateCount, e
}

func (state *billingState) validateBillLineItem(ctv workflow.Context, lineItem model.BillLineItem) error {
	state.logger.Info("Validating bill line item", "Bill", state.BillInfo, "Line item", lineItem)
	return state.BillInfo.CheckLineItemCompatible(lineItem)
}

func (state *billingState) addBillLineItemIfNotExistSyncActivity(ctx workflow.Context, lineItem model.BillLineItem) (intermediateState BillingState, e error) {
	state.logger.Info("Adding bill line item if it does not exist", "Bill", state.BillInfo, "Line item", lineItem)
	ctxWithOptions := workflow.WithActivityOptions(ctx, defaultActivityOptions())
	var updateCount uint64
	e = workflow.ExecuteActivity(
		ctxWithOptions,
		activity.AddBillLineItemIfNotExistActivity,
		state.BillInfo,
		lineItem,
	).Get(ctxWithOptions, &updateCount)
	if e == nil && 0 < updateCount {
		state.BillLineItemCount += updateCount
		state.Total.Add(lineItem.Amount)
		state.logger.Info("Bill line item added", "Total", state.Total, "Amount", lineItem.Amount)
	}
	return state.Clone(), e
}

func (state *billingState) closeBillSyncActivity(ctx workflow.Context) (uint64, error) {
	state.logger.Info("Bill line items workflow completed", "Bill", state.BillInfo, "Final count value", state.BillLineItemCount)
	ctxWithOptions := workflow.WithActivityOptions(ctx, defaultActivityOptions())
	var updateCount uint64
	e := workflow.ExecuteActivity(ctxWithOptions, activity.CloseBillActivity, state.BillInfo).Get(ctxWithOptions, &updateCount)
	return updateCount, e
}

func BillingWorkflow(ctx workflow.Context, billInfo model.BillInfo, duration time.Duration) (count BillingState, e error) {
	state := &billingState{
		BillingState: BillingState{
			BillInfo:          billInfo,
			BillLineItemCount: 0,
			Total: TotalAmount{
				Total: model.Amount{Number: 0, CurrencyCode: billInfo.CurrencyCode},
				Ok:    true,
			},
		},
		logger: workflow.GetLogger(ctx),
	}
	state.logger.Info("Bill line items workflow started", "Bill", billInfo, "Duration", duration)

	if duration < 0 {
		return state.Clone(), NegativeDurationError{duration}
	}

	if _, e := state.createBillIfNotExistSyncActivity(ctx); e != nil {
		return state.Clone(), e
	}

	e = workflow.SetUpdateHandlerWithOptions(
		ctx,
		AddBillLineItemUpdate,
		state.addBillLineItemIfNotExistSyncActivity,
		workflow.UpdateHandlerOptions{
			Validator: state.validateBillLineItem,
		})
	if e != nil {
		return state.Clone(), e
	}
	e = workflow.SetQueryHandler(ctx, GetPendingBillStateQuery, func() (BillingState, error) {
		return state.Clone(), nil
	})
	if e != nil {
		return state.Clone(), e
	}

	// Create a selector to either end with timer or close the bill ahead of time
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(
		workflow.NewTimer(ctx, duration),
		func(future workflow.Future) {
			state.logger.Info("Bill arrived at maturity, closing")
		})
	selector.AddReceive(
		workflow.GetSignalChannel(ctx, CloseBillEarlySignal),
		func(channel workflow.ReceiveChannel, more bool) {
			var receivedUpdate string
			channel.Receive(ctx, &receivedUpdate)
			state.logger.Info("Received signal to close bill early:", receivedUpdate)
		})
	selector.Select(ctx) // Wait until either the timer expires or the close signal is received

	_, e = state.closeBillSyncActivity(ctx)
	if e == nil {
		state.BillInfo.Status = model.Closed
	}
	return state.Clone(), e
}
