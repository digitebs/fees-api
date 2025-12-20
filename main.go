package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"fees-api/activities"
	"fees-api/workflow"
)

func main() {
	// Create a Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	defer c.Close()

	// Create a Temporal worker for your billing task queue
	w := worker.New(c, "BILLING_TASK_QUEUE", worker.Options{})

	// Register your workflow and activities with the worker
	w.RegisterWorkflow(workflow.BillWorkflow)
	w.RegisterActivity(activities.FinalizeBillActivity)
	w.RegisterActivity(activities.PersistLineItemActivity)

	// Start listening to the task queue in a separate goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Println("Starting Temporal worker on task queue BILLING_TASK_QUEUE...")
		errCh <- w.Run(worker.InterruptCh())
	}()

	// Wait for termination signal (CTRL+C or SIGTERM)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("Received signal %s: shutting down worker...\n", sig)
		w.Stop() // graceful shutdown
	case err := <-errCh:
		if err != nil {
			log.Fatalf("Worker stopped with error: %v", err)
		}
	}

	log.Println("Worker shutdown complete.")
}
