package db

import (
	"coding-challenge/pkg/model"
	"database/sql"
	"fmt"
)

const SqlDbType = "sql"

type SqlBillDatabase struct {
	sql *sql.DB
}

var _ BillDatabase = SqlBillDatabase{}

func NewSqlBillDatabase(sql *sql.DB) *SqlBillDatabase {
	return &SqlBillDatabase{
		sql: sql,
	}
}

func (m SqlBillDatabase) CreateBill(bill model.BillInfo) (uint64, error) {
	res, err := m.sql.Exec(`
		INSERT INTO Bill (CustomerId, Id, CurrencyCode)
		VALUES ($1, $2, $3)
		ON CONFLICT (CustomerId, Id) DO NOTHING;
	`, string(bill.Id.CustomerId),
		bill.Id.Id,
		bill.CurrencyCode)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	fmt.Printf("Sql saving bill: %v, rows %d\n", bill, rowsAffected)
	return uint64(rowsAffected), nil
}

func (m SqlBillDatabase) AddLineItem(lineItem model.BillLineItem, totalBefore model.TotalAmount) (uint64, error) {
	tx, err := m.sql.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`
		SELECT Status, CurrencyCode
		FROM Bill
		WHERE CustomerId = $1 AND Id = $2;
	`, string(lineItem.Id.BillId.CustomerId), lineItem.Id.BillId.Id)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, ErrBillNotFound
	}
	var status model.BillStatus
	var currencyCode string
	err = rows.Scan(&status, &currencyCode)
	if err != nil {
		return 0, err
	}
	if status != model.Open {
		return 0, ErrBillClosed
	}
	if lineItem.Amount.CurrencyCode != model.CurrencyCode(currencyCode) {
		return 0, ErrCurrencyMismatch
	}
	rows.Close()
	totalBefore.Add(lineItem.Amount)
	res, err := tx.Exec(`
		UPDATE Bill
		SET
			LineItemCount = LineItemCount + 1,
			TotalAmount = $3,
			TotalOk = $4
		WHERE CustomerId = $1 AND Id = $2;
	`, string(lineItem.Id.BillId.CustomerId),
		lineItem.Id.BillId.Id,
		totalBefore.Total.Number,
		totalBefore.Ok)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrBillNotFound
	}
	res, err = tx.Exec(`
		INSERT INTO LineItem (CustomerId, BillId, Id, Description, Amount)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (CustomerId, BillId, Id) DO NOTHING;
	`, string(lineItem.Id.BillId.CustomerId),
		lineItem.Id.BillId.Id,
		lineItem.Id.Id,
		lineItem.Description,
		lineItem.Amount.Number)
	if err != nil {
		return 0, err
	}
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}
	fmt.Printf("Sql saving lineItem: %v, rows %d\n", lineItem, rowsAffected)
	if rowsAffected == 0 {
		return 0, nil
	}
	return uint64(rowsAffected), tx.Commit()
}

func (m SqlBillDatabase) CloseBill(billId model.BillId) (uint64, error) {
	res, err := m.sql.Exec(`
		UPDATE Bill
		SET Status = $3
		WHERE CustomerId = $1 AND Id = $2;
	`, string(billId.CustomerId), billId.Id, model.Closed)
	fmt.Printf("Sql Closing: %v\n", billId)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrBillNotFound
	}
	return uint64(rowsAffected), nil
}

func (m SqlBillDatabase) GetBill(billId model.BillId) (BillInfoAndMetadata, error) {
	rows, err := m.sql.Query(`
		SELECT CustomerId, Id, Status, LineItemCount, TotalAmount, TotalOk, CurrencyCode
		FROM Bill
		WHERE CustomerId = $1 AND Id = $2;
	`, string(billId.CustomerId), billId.Id)
	if err != nil {
		return BillInfoAndMetadata{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return BillInfoAndMetadata{}, ErrBillNotFound
	}
	var (
		customerId    string
		id            string
		status        model.BillStatus
		lineItemCount uint64
		totalAmount   int64
		totalOk       bool
		currencyCode  string
	)
	err = rows.Scan(&customerId, &id, &status, &lineItemCount, &totalAmount, &totalOk, &currencyCode)
	if err != nil {
		return BillInfoAndMetadata{}, err
	}
	return BillInfoAndMetadata{
		BillInfo: model.BillInfo{
			Id: model.BillId{
				CustomerId: model.CustomerId(customerId),
				Id:         id,
			},
			Status:       status,
			CurrencyCode: model.CurrencyCode(currencyCode),
		},
		LineItemCount: lineItemCount,
		TotalAmount:   model.Amount{Number: totalAmount, CurrencyCode: model.CurrencyCode(currencyCode)},
		TotalOk:       totalOk,
	}, nil
}
