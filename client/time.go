package client

import "time"

const nanoToMilliseconds = 1000 * 1000

// timestamp
func (client *Instance) Now() uint64 {
	// @todo correction for drift
	return uint64(time.Now().UnixNano() / nanoToMilliseconds)
}
