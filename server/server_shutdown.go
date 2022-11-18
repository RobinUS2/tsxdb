package server

import (
	"log"
	"sync/atomic"
	"time"
)

// stop listening, this should only be called once per instance
func (instance *Instance) Shutdown() error {
	log.Println("shutting down")
	atomic.StoreInt32(&instance.shuttingDown, 1)

	// tickers
	instance.statsTicker.Stop()
	instance.sessionTicker.Stop()

	// poll RPC listener shutdown
	if instance.RpcListener() != nil {
		// pending
		v := atomic.LoadInt64(&instance.pendingRequests)
		if v > 0 {
			// 50 x 100ms => 5 second max
			for i := 0; i < 50; i++ {
				time.Sleep(100 * time.Millisecond)
				v := atomic.LoadInt64(&instance.pendingRequests)
				if instance.RpcListener() == nil || v == 0 {
					break
				}
			}
		}
		// force shutdown
		if err := instance.closeRpc(); err != nil {
			return err
		}
	}

	// shutdown telnet
	if instance.telnetServer != nil {
		if err := instance.telnetServer.Shutdown(); err != nil {
			return err
		}
	}

	log.Println("shutdown complete")
	return nil
}
