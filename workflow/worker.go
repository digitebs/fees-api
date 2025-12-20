package workflow

import (
	"log"

	"fees-api/bill"

	"go.temporal.io/sdk/worker"
)

const taskQueue = "billing_task_queue"

func init() {
	w := worker.New(bill.GetTemporalClient(), "BILLING_TASK_QUEUE", worker.Options{})

	// Register your workflow and activities with the worker
	w.RegisterWorkflow(BillWorkflow)
	w.RegisterActivity(FinalizeBillActivity)
	w.RegisterActivity(PersistLineItemActivity)

	// Start listening to the task queue in a separate goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Println("Starting Temporal worker on task queue BILLING_TASK_QUEUE...")
		errCh <- w.Run(worker.InterruptCh())
	}()
}
