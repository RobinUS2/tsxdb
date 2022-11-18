package integration

import (
	"errors"
	"github.com/RobinUS2/tsxdb/client"
)

func Run() error {
	opts := client.NewOpts()
	if opts == nil {
		return errors.New("missing opts")
	}
	return nil
}
