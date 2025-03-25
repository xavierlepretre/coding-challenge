package model

type BillId struct {
	CustomerId CustomerId
	Id         string
}

type BillStatus uint8

const (
	Open BillStatus = iota
	Closed
)

type BillInfo struct {
	Id           BillId
	CurrencyCode CurrencyCode
	Status       BillStatus
}

func (b *BillInfo) CheckLineItemCompatible(lineItem BillLineItem) error {
	return CheckCurrencyCodeCompatible(b.CurrencyCode, lineItem.Amount.CurrencyCode)
}

type Bill struct {
	Info      BillInfo
	LineItems []BillLineItem
}

func NewBill(info BillInfo, lineItems []BillLineItem) Bill {
	return NewBillWithCapacity(info, lineItems, len(lineItems))
}

func NewBillWithCapacity(info BillInfo, lineItems []BillLineItem, capacity int) Bill {
	bill := Bill{Info: info, LineItems: make([]BillLineItem, len(lineItems), capacity)}
	copy(bill.LineItems, lineItems)
	return bill
}

func (b *Bill) GetLineItemsCopy() []BillLineItem {
	lineItems := make([]BillLineItem, len(b.LineItems))
	copy(lineItems, b.LineItems)
	return lineItems
}

func (b *Bill) GetLineItemsCount() int {
	return len(b.LineItems)
}

func (b *Bill) Clone() Bill {
	return NewBillWithCapacity(b.Info, b.LineItems, len(b.LineItems))
}

type BillLineItemId struct {
	BillId BillId
	Id     string
}

type BillLineItem struct {
	Id          BillLineItemId
	Description string
	Amount      Amount
}

func (b *Bill) AddLineItem(lineItem BillLineItem) error {
	if e := b.Info.CheckLineItemCompatible(lineItem); e != nil {
		return e
	}
	b.LineItems = append(b.LineItems, lineItem)
	return nil
}

type TotalAmount struct {
	Total Amount
	Ok    bool
}

func (total *TotalAmount) Add(amount Amount) {
	if total.Ok {
		total.Total, total.Ok = total.Total.Add(amount)
	}
}
