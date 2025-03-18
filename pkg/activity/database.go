package activity

import (
	"coding-challenge/pkg/model"
	"fmt"
	"time"
)

const SaveToDatabaseActivityTimeout = time.Second

func createBillIfNotExistInDatabaseActivity(bill model.BillInfo) error {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Saving: %v", bill) // Replace with DB write operation
	return nil
}

func addBillLineItemIfNotExistToDatabaseActivity(bill model.BillInfo, lineItem model.BillLineItem) error {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Saving: %v", lineItem) // Replace with DB write operation
	return nil
}

func closeBillInDatabaseActivity(bill model.BillInfo) error {
	// Simulate storing data (Replace with actual DB logic)
	// Example: Write to a SQL database, Redis, or S3
	fmt.Printf("Closing: %v", bill) // Replace with DB write operation
	return nil
}
