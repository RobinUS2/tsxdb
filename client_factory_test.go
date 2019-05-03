package tsxdb_test

import (
	"../tsxdb"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := tsxdb.NewClient(tsxdb.NewClientOpts())
	if c == nil {
		t.Error(c)
	}
}
