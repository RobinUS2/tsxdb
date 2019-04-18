package tools

import (
	"crypto/hmac"
	"crypto/sha512"
	"github.com/OneOfOne/xxhash"
	"strconv"
)

func HmacInt(secret []byte, v int) int {
	signature, _ := Hmac(secret, []byte(strconv.Itoa(v)))
	h := xxhash.New32()
	if _, err := h.Write([]byte(signature)); err != nil {
		panic(err)
	}
	return int(h.Sum32())
}

func Hmac(secret []byte, b []byte) (signature []byte, err error) {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(b)
	signature = mac.Sum(nil)
	return
}
