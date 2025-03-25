package db

import (
	"coding-challenge/pkg/model"
	"fmt"
	"sync"
)

type storedBillAndItems struct {
	bill          model.BillInfo
	lineItems     map[string]*model.BillLineItem
	lineItemCount uint64
	totalAmount   model.Amount
	totalOk       bool
}

type customerBills struct {
	bills map[string]*storedBillAndItems
}

type InMemoryBillDatabase struct {
	// customerId -> Id -> bill info
	bills map[model.CustomerId]*customerBills
	mu    *sync.RWMutex
}

var _ BillDatabase = InMemoryBillDatabase{}

func NewInMemoryBillDatabase() *InMemoryBillDatabase {
	return &InMemoryBillDatabase{
		bills: make(map[model.CustomerId]*customerBills),
		mu:    &sync.RWMutex{},
	}
}

func (m InMemoryBillDatabase) CreateBill(bill model.BillInfo) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	customerId, billId := bill.Id.CustomerId, bill.Id.Id
	if _, ok := m.bills[customerId]; !ok {
		m.bills[customerId] = &customerBills{
			bills: make(map[string]*storedBillAndItems),
		}
	} else if _, ok := m.bills[customerId].bills[billId]; ok {
		return 0, ErrBillAlreadyExists
	}

	m.bills[customerId].bills[billId] = &storedBillAndItems{
		bill:      bill,
		lineItems: make(map[string]*model.BillLineItem),
	}
	fmt.Printf("In Memory Saving: %v\n", bill)
	return 1, nil
}

func (m InMemoryBillDatabase) AddLineItem(lineItem model.BillLineItem, _ model.TotalAmount) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	customerId, billId, lineItemId := lineItem.Id.BillId.CustomerId, lineItem.Id.BillId.Id, lineItem.Id.Id
	customerBills, ok := m.bills[customerId]
	if !ok {
		return 0, ErrBillNotFound
	}
	storedBill, ok := customerBills.bills[billId]
	if !ok {
		return 0, ErrBillNotFound
	}
	if storedBill.bill.Status == model.Closed {
		return 0, ErrBillClosed
	}
	if _, ok := storedBill.lineItems[lineItemId]; ok {
		return 0, ErrLineItemAlreadyExists
	}

	storedBill.lineItems[lineItemId] = &lineItem
	storedBill.lineItemCount++
	if storedBill.totalOk {
		storedBill.totalAmount, storedBill.totalOk = storedBill.totalAmount.Add(lineItem.Amount)
	}
	fmt.Printf("In Memory Saving: %v\n", lineItem)
	return 1, nil
}

func (m InMemoryBillDatabase) CloseBill(billId model.BillId) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	customerId, id := billId.CustomerId, billId.Id
	customerBills, ok := m.bills[customerId]
	if !ok {
		return 0, ErrBillNotFound
	}
	storedBillAndItems, ok := customerBills.bills[id]
	if !ok {
		return 0, ErrBillNotFound
	}

	storedBillAndItems.bill.Status = model.Closed
	customerBills.bills[id] = storedBillAndItems
	fmt.Printf("In Memory Closing: %v\n", billId)
	return 1, nil
}

func (m InMemoryBillDatabase) GetBill(billId model.BillId) (BillInfoAndMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	customerId, id := billId.CustomerId, billId.Id
	customerBills, ok := m.bills[customerId]
	if !ok {
		return BillInfoAndMetadata{}, ErrBillNotFound
	}
	storedBillAndItems, ok := customerBills.bills[id]
	if !ok {
		return BillInfoAndMetadata{}, ErrBillNotFound
	}

	return BillInfoAndMetadata{
		BillInfo:      storedBillAndItems.bill,
		LineItemCount: storedBillAndItems.lineItemCount,
		TotalAmount:   storedBillAndItems.totalAmount,
		TotalOk:       storedBillAndItems.totalOk,
	}, nil
}
