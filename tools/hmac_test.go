package tools_test

import (
	"github.com/RobinUS2/tsxdb/tools"
	"testing"
)

func TestHmac(t *testing.T) {
	var secret = []byte("test")
	var msg = []byte("msg")
	signature, err := tools.Hmac(secret, msg)
	if err != nil {
		t.Error(err)
	}
	if signature == nil || len(signature) < 1 {
		t.Error()
	}
}
func TestHmacInt(t *testing.T) {
	var secret = []byte("test")
	signature := tools.HmacInt(secret, 1234)
	if signature == 0 {
		t.Error()
	}
}
