package activity

import (
	"coding-challenge/pkg/model"
)

func CreateBillIfNotExistActivity(bill model.BillInfo) error {
	return createBillIfNotExistInDatabaseActivity(bill)
}

func AddBillLineItemIfNotExistActivity(bill model.BillInfo, lineItem model.BillLineItem) error {
	return addBillLineItemIfNotExistToDatabaseActivity(bill, lineItem)
}

func CloseBillActivity(bill model.BillInfo) error {
	return closeBillInDatabaseActivity(bill)
}
