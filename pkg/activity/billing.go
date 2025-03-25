package activity

import (
	"coding-challenge/pkg/model"
)

type ActivityHost interface {
	CreateBillIfNotExistActivity(bill model.BillInfo) (uint64, error)
	AddBillLineItemIfNotExistActivity(lineItem model.BillLineItem, totalBefore model.TotalAmount) (uint64, error)
	CloseBillActivity(bill model.BillInfo) (uint64, error)
}

type DummyActivityHost struct {
}

var _ ActivityHost = &DummyActivityHost{}

func (d *DummyActivityHost) CreateBillIfNotExistActivity(bill model.BillInfo) (uint64, error) {
	panic("Not implemented")
}

func (d *DummyActivityHost) AddBillLineItemIfNotExistActivity(lineItem model.BillLineItem, totalBefore model.TotalAmount) (uint64, error) {
	panic("Not implemented")
}

func (d *DummyActivityHost) CloseBillActivity(bill model.BillInfo) (uint64, error) {
	panic("Not implemented")
}
