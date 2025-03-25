package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBillIdWorksAsId(t *testing.T) {
	// Arrange
	billId1 := BillId{CustomerId: "alice", Id: "6c5bb10f-6fd2-49be-a75a-806ad1c4cfcf"}
	billId2 := BillId{CustomerId: "alice", Id: "6c5bb10f-6fd2-49be-a75a-806ad1c4cfcf"}
	billId3 := BillId{CustomerId: "alice", Id: "6c5bb10f-6fd2-49be-a75a-806ad1c4cfce"} // Slight change in Id
	billId4 := BillId{CustomerId: "alicf", Id: "6c5bb10f-6fd2-49be-a75a-806ad1c4cfcf"} // Slight change in CustomerId

	// Act nil

	// Assert
	assert.Equal(t, billId1, billId2)
	assert.True(t, billId1 == billId2)
	assert.NotEqual(t, billId1, billId3)
	assert.True(t, billId1 != billId3)
	assert.NotEqual(t, billId1, billId4)
	assert.True(t, billId1 != billId4)
}

func TestAddLineItemOnEmpty(t *testing.T) {
	// Arrange
	billId := BillId{CustomerId: "alice", Id: "6c5bb10f-6fd2-49be-a75a-806ad1c4cfcf"}
	billInfo := BillInfo{Id: billId, CurrencyCode: "USD", Status: Open}
	bill := NewBillWithCapacity(billInfo, []BillLineItem{}, 1)
	one, e := NewAmountFromInt64(100, "USD")
	assert.NoError(t, e)
	matchBoxId := BillLineItemId{BillId: billId, Id: "dfddc86a-6e07-4476-9b53-4ae78b4bce1c"}
	matchboxItem := BillLineItem{Id: matchBoxId, Description: "Matchbox", Amount: one}

	// Act
	e = bill.AddLineItem(matchboxItem)

	// Assert
	assert.NoError(t, e)
	assert.Len(t, bill.LineItems, 1)
	assert.EqualValues(t, matchboxItem, bill.LineItems[0])
}

func TestAddLineItemOnNonEmpty(t *testing.T) {
	// Arrange
	one, e := NewAmountFromInt64(100, "USD")
	assert.NoError(t, e)
	billId := BillId{CustomerId: "bob", Id: "91c05476-2ae1-4fcf-a25c-f1851847aafe"}
	matchboxId1 := BillLineItemId{BillId: billId, Id: "eef5a9e9-0d64-440f-87b3-3e7c02910d1f"}
	matchboxItem1 := BillLineItem{Id: matchboxId1, Description: "Matchbox", Amount: one}
	billInfo := BillInfo{Id: billId, CurrencyCode: "USD", Status: Open}
	bill := NewBillWithCapacity(billInfo, []BillLineItem{matchboxItem1}, 3)
	two, e := NewAmountFromInt64(200, "USD")
	assert.NoError(t, e)
	candleId2 := BillLineItemId{BillId: billId, Id: "3ab47a6a-2563-4c4e-a963-8bf07f10d52a"}
	candleItem2 := BillLineItem{Id: candleId2, Description: "Candle", Amount: two}

	// Act
	e = bill.AddLineItem(candleItem2)

	// Assert
	assert.NoError(t, e)
	assert.Len(t, bill.LineItems, 2)
	assert.EqualValues(t, matchboxItem1, bill.LineItems[0])
	assert.EqualValues(t, candleItem2, bill.LineItems[1])
}

func TestAddLineItemWithDifferentCurrencyCode(t *testing.T) {
	// Arrange
	one, e := NewAmountFromInt64(100, "USD")
	assert.NoError(t, e)
	billId := BillId{CustomerId: "carol", Id: "fc03932f-2b53-4d07-ad55-24fc7d85e277"}
	matchboxId1 := BillLineItemId{BillId: billId, Id: "e29ed54f-5827-4286-bcf7-777f346e1039"}
	matchboxItem1 := BillLineItem{Id: matchboxId1, Description: "Matchbox", Amount: one}
	billInfo := BillInfo{Id: billId, CurrencyCode: "USD", Status: Open}
	bill := NewBillWithCapacity(billInfo, []BillLineItem{matchboxItem1}, 3)
	two, e := NewAmountFromInt64(200, "GEL")
	assert.NoError(t, e)
	candleId2 := BillLineItemId{BillId: billId, Id: "6f3d248e-f460-47fe-82dd-f2115a35e5ac"}
	candleItem2 := BillLineItem{Id: candleId2, Description: "Candle", Amount: two}

	// Act
	e = bill.AddLineItem(candleItem2)

	// Assert
	assert.ErrorIs(t, e, IncompatibleCurrencyCodesError{"USD", "GEL"})
	assert.Len(t, bill.LineItems, 1)
	assert.EqualValues(t, matchboxItem1, bill.LineItems[0])
}
