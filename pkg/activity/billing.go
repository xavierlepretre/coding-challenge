package activity

import (
	"coding-challenge/pkg/model"
)

func CreateBillIfNotExistActivity(bill model.BillInfo) (uint64, error) {
	return createBillIfNotExistInDatabaseActivity(bill)
}

func AddBillLineItemIfNotExistActivity(bill model.BillInfo, lineItem model.BillLineItem) (uint64, error) {
	return addBillLineItemIfNotExistToDatabaseActivity(bill, lineItem)
}

func CloseBillActivity(bill model.BillInfo) (uint64, error) {
	return closeBillInDatabaseActivity(bill)
}
