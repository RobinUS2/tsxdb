package tools_test

import (
	"github.com/RobinUS2/tsxdb/tools"
	"testing"
)

func TestRandomInsecureIdentifier(t *testing.T) {
	if tools.RandomInsecureIdentifier() == 0 {
		t.Error()
	}
}
