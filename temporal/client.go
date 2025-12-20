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

func InitClient(addr string) {
	once.Do(func() {
		c, err := client.Dial(client.Options{HostPort: addr})
		if err != nil {
			log.Fatalf("unable to create Temporal client: %v", err)
		}
		Client = c
	})
}

func GetClient() client.Client {
	return Client
}
