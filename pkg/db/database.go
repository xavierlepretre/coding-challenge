package db

import (
	"coding-challenge/pkg/model"
	"errors"
)

type BillInfoAndMetadata struct {
	BillInfo      model.BillInfo
	LineItemCount uint64
	TotalAmount   model.Amount
	TotalOk       bool
}

type BillDatabase interface {
	CreateBill(bill model.BillInfo) (uint64, error)
	AddLineItem(lineItem model.BillLineItem, totalBefore model.TotalAmount) (uint64, error)
	CloseBill(billId model.BillId) (uint64, error)
	GetBill(billId model.BillId) (BillInfoAndMetadata, error)
}

// ErrBillNotFound is returned when a bill is not found.
var ErrBillNotFound = errors.New("bill not found")

// ErrBillAlreadyExists is returned when a bill already exists.
var ErrBillAlreadyExists = errors.New("bill already exists")

// ErrLineItemAlreadyExists is returned when a line item already exists.
var ErrLineItemAlreadyExists = errors.New("line item already exists")

// ErrBillClosed is returned when a bill is closed.
var ErrBillClosed = errors.New("bill is closed")

// ErrBillMismatch is returned when a line item is expected to but does not belong to a bill
var ErrBillMismatch = errors.New("bill and lineItem mismatch")

// ErrCurrencyMismatch is returned when a line item and bill have mismatched currency codes
var ErrCurrencyMismatch = errors.New("bill and lineItem have mismatched currency code")
