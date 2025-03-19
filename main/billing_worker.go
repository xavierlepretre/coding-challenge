package main

import (
	"coding-challenge/pkg/activity"
	"coding-challenge/pkg/workflow"
	"flag"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const billingQueueFlag = "task-queue"

func main() {
	// Define a flag for the task queue
	taskQueue := flag.String(billingQueueFlag, workflow.BillingQueueDefault, "Specify the billing task queue name")
	flag.Parse()

	fmt.Printf("Starting worker for task queue: %s\n", *taskQueue)

	// Create Temporal client
	client, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("unable to create Temporal client: %v", err)
	}
	defer client.Close()

	// Create a worker for a specific task queue
	w := worker.New(client, *taskQueue, worker.Options{})

	// Register your workflow and activities
	w.RegisterWorkflow(workflow.BillingWorkflow)
	w.RegisterActivity(activity.CreateBillIfNotExistActivity)
	w.RegisterActivity(activity.AddBillLineItemIfNotExistActivity)
	w.RegisterActivity(activity.CloseBillActivity)

	// Start the worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("unable to start worker: %v", err)
	}
}
