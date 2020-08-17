package server

import (
	"log"
	"sync"

	"github.com/iostrovok/conveyormaster/server/grpc"
	"github.com/iostrovok/conveyormaster/server/http"
	"github.com/iostrovok/conveyormaster/server/messager"
)

// Just returns first nonempty string
func getNoEmptyString(in ...string) string {
	for _, k := range in {
		if k != "" {
			return k
		}
	}

	return ""
}

func Start() {
	grpcAddr := "127.0.0.1:5100"
	httpAddr := "127.0.0.1:5000"

	var wg sync.WaitGroup

	message := messager.New()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Fatal(grpc.Start(grpcAddr, message))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Fatal(http.Start(httpAddr, message))
	}()

	wg.Wait()
}
