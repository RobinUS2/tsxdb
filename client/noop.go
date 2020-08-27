package client

import (
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/pkg/errors"
)

func (series *Series) NoOp() error {

	// get
	conn, err := series.client.GetConnection()
	if err != nil {
		return err
	}
	defer panicOnErrorClose(conn.Close)

	// session data
	return handleRetry(func() error {
		request := types.NoOpRequest{}
		request.SessionTicket = conn.getSessionTicket()

		// execute
		var response *types.NoOpResponse
		if err := conn.client.Call(types.EndpointNoOp.String()+"."+types.MethodName, request, &response); err != nil {
			return err
		}
		if response.Error != nil {
			return errors.New(response.Error.Error().Error())
		}
		return nil
	})
}
