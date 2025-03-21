package model

import "github.com/google/uuid"

type BillIdGenerator interface {
	New() string
}

type UuidBillIdGenerator struct{}

var _ BillIdGenerator = &UuidBillIdGenerator{}

func (*UuidBillIdGenerator) New() string {
	return uuid.New().String()
}
