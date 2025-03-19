package activity

import (
	"coding-challenge/pkg/model"
	"fmt"
	"time"
)

const SaveToDatabaseActivityTimeout = time.Second

func createBillIfNotExistInDatabaseActivity(bill model.BillInfo) (uint64, error) {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Saving: %v", bill) // Replace with DB write operation
	return 1, nil
}

func addBillLineItemIfNotExistToDatabaseActivity(bill model.BillInfo, lineItem model.BillLineItem) (uint64, error) {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Saving: %v", lineItem) // Replace with DB write operation
	return 1, nil
}

func closeBillInDatabaseActivity(bill model.BillInfo) (uint64, error) {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Closing: %v", bill) // Replace with DB write operation
	return 1, nil
}
