package client_test

import (
	"github.com/Route42/tsxdb/client"
	"testing"
)

func TestNewShallowImportClient(t *testing.T) {
	c := client.DefaultClient()
	if c == nil {
		t.Error()
	}
}
