package shutdown_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/eloylp/kit/shutdown"
)

func ExampleWithOSSignals() {
	s := &http.Server{
		Addr: "0.0.0.0:8080",
	}
	wg := &sync.WaitGroup{}
	shutdown.WithOSSignals(s, 5*time.Second, wg, func(err error) { // We can also pass nil as errHandler func if we dont care about errors.
		fmt.Printf("shutdown error: %v \n", err)
	})
	// Send interrupt signal in two seconds
	go func() {
		<-time.NewTimer(2 * time.Second).C
		p, _ := os.FindProcess(os.Getpid()) // As this is an example, errors are omitted.
		_ = p.Signal(os.Interrupt)
	}()
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	wg.Wait() // Ensure waiting for shutdown to return
	fmt.Println("Waited for shutdown function, so we end the program.")
	// Output:
	// Waited for shutdown function, so we end the program.
}
