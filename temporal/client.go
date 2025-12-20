package temporal

import (
	"log"
	"sync"

	"go.temporal.io/sdk/client"
)

var (
	Client client.Client
	once   sync.Once
)

// GetClient returns the Temporal client, initializing it with the provided server address.
func GetClient(serverAddr string) client.Client {
	once.Do(func() {
		c, err := client.Dial(client.Options{HostPort: serverAddr})
		if err != nil {
			log.Fatalf("unable to create Temporal client: %v", err)
		}
		Client = c
	})
	return Client
}
