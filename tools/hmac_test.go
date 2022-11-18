package tools_test

import (
	"fmt"
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
	const expected = "5cc3c04381b3100c65a20540f17c1917676c5813a4ce0a7d0281f69e56775035c746da69f531bb8b13c57172aeb7e9d25fb18cecd070369c99ec7e49f7034b69"
	res := fmt.Sprintf("%x", signature)
	if res != expected {
		t.Error(res, expected)
	}
}

func TestHmacInt(t *testing.T) {
	var secret = []byte("test")
	signature := tools.HmacInt(secret, 1234)
	if signature != 4017713220 {
		t.Error(signature)
	}
}
