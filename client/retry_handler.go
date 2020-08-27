package client

import (
	"log"
	"math/rand"
	"runtime/debug"
	"time"
)

const DefaultRpcBaseSleep = 100 * time.Millisecond

func jitterSleep(baseSleep time.Duration, attempt int) {
	if attempt == 0 {
		return
	}
	time.Sleep((baseSleep * time.Duration(attempt*attempt) * time.Millisecond) + time.Duration(rand.Intn(50)))
}

func handleRetry(fn func() error) (err error) {
	defer func() {
		// recover from panics, way to signal stop retrying
		if r := recover(); r != nil {
			log.Printf("recovered retry %s", r) // @todo proper way to signal non-retryable errors from handleRetry
			err = nil
		}
	}()

	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		jitterSleep(DefaultRpcBaseSleep, i)
		err = fn()
		if err != nil {
			log.Printf("failed attempt: %s", err)
			debug.PrintStack()
			continue
		}
		break
	}
	return err
}
