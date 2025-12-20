package workflow

import (
	"context"
	"log"

	"fees-api/bill"

	"go.temporal.io/sdk/worker"
)

const taskQueue = "billing_task_queue"

//encore:service
type Service struct{}

// initService initializes the Temporal worker when the Encore service starts
func initService() (*Service, error) {
	log.Println("Initializing Temporal workflow service...")

	w := worker.New(bill.GetTemporalClient(), "BILLING_TASK_QUEUE", worker.Options{})

	// Register your workflow and activities with the worker
	w.RegisterWorkflow(BillWorkflow)
	w.RegisterActivity(FinalizeBillActivity)
	w.RegisterActivity(PersistLineItemActivity)

	// Start listening to the task queue in a separate goroutine
	go func() {
		log.Println("Starting Temporal worker on task queue BILLING_TASK_QUEUE...")
		if err := w.Run(worker.InterruptCh()); err != nil {
			log.Printf("Temporal worker error: %v", err)
		}
	}()

	return &Service{}, nil
}

// HealthCheck provides a way to verify the workflow service is running
//
//encore:api private
func (s *Service) HealthCheck(ctx context.Context) error {
	// Could add actual health checks here
	return nil
}
