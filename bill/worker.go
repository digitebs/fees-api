package bill

import (
	"log"

	"go.temporal.io/sdk/worker"
)

const taskQueue = "BILLING_TASK_QUEUE"

// initService initializes the Temporal worker when the Encore service starts
func initService() (*Service, error) {
	log.Println("Initializing Temporal workflow service...")

	w := worker.New(GetTemporalClient(), taskQueue, worker.Options{})

	// Register your workflow and activities with the worker
	w.RegisterWorkflow(BillWorkflow)
	w.RegisterActivity(FinalizeBillActivity)
	w.RegisterActivity(AddLineItemActivity)

	// Start listening to the task queue in a separate goroutine
	go func() {
		log.Println("Starting Temporal worker on task queue BILLING_TASK_QUEUE...")
		if err := w.Run(worker.InterruptCh()); err != nil {
			log.Printf("Temporal worker error: %v", err)
		}
	}()

	return &Service{}, nil
}
